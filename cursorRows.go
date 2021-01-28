// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/SAP/go-dblib/asetypes"
	"github.com/SAP/go-dblib/tds"
)

var (
	_ driver.Rows                           = (*CursorRows)(nil)
	_ driver.RowsColumnTypeLength           = (*CursorRows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*CursorRows)(nil)

	ErrCurNoMoreRows = errors.New("no more rows in cursor")
)

type CursorRows struct {
	cursor *Cursor
	rows   chan *tds.RowPackage
	// TODO this is just a workaround
	rowsClosed bool

	// The count of read rows is required to update/delete a row through
	// the cursor.
	readRows  int
	totalRows int
}

func (cursor *Cursor) NewCursorRows() (*CursorRows, error) {
	return &CursorRows{
		cursor: cursor,
		rows:   make(chan *tds.RowPackage, cursor.conn.Info.CursorCacheRows),
	}, nil
}

func (rows *CursorRows) Close() error {
	return rows.cursor.Close(context.Background())
}

func (rows CursorRows) Columns() []string {
	// TODO ignore hidden columns
	response := make([]string, len(rows.cursor.rowFmt.Fmts))

	for i, fieldFmt := range rows.cursor.rowFmt.Fmts {
		// TODO check if RowFmt is wide and contains column label,
		// catalgoue, schema, table
		response[i] = fieldFmt.Name()
	}

	return response
}

func (rows *CursorRows) Next(dst []driver.Value) error {
	rowPkg, err := rows.nextPkg(context.Background())
	if err != nil {
		// Signal io.EOF to database/sql if no more rows can be read
		if errors.Is(err, ErrCurNoMoreRows) {
			return io.EOF
		}
		return fmt.Errorf("go-ase: error getting next row: %w", err)
	}

	for i := range dst {
		dst[i] = rowPkg.DataFields[i].Value()
	}
	rows.readRows++

	return nil
}

func (rows *CursorRows) nextPkg(ctx context.Context) (*tds.RowPackage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case rowPkg, ok := <-rows.rows:
		if ok {
			return rowPkg, nil
		}
	default:
	}

	// fetch more rows
	if err := rows.fetch(ctx); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("error fetching more rows: %w", err)
	}

	return rows.nextPkg(ctx)
}

func (rows *CursorRows) fetch(ctx context.Context) error {
	// Set the last received package to the rowfmt received during
	// setup. The params/rows packages need the information from the
	// format to setup the data fields.
	rows.cursor.conn.Channel.SetLastPkgRx(rows.cursor.rowFmt)

	fetchPkg := &tds.CurFetchPackage{
		CursorID: rows.cursor.cursorID,
		Name:     rows.cursor.name,
		Type:     tds.TDS_CUR_NEXT,
	}
	if err := rows.cursor.conn.Channel.SendPackage(ctx, fetchPkg); err != nil {
		return fmt.Errorf("error sending CurFetchPackage: %w", err)
	}

	// This is a really ugly workaround.
	// CurInfo doesn't report the total rows on scrollable cursors. But
	// it should report the TotalRows once the cursor scrolled past the
	// last row in the data set - but at least ASE doesn't.
	// TDS also specifies that CurInfo should report RowNum accurately,
	// which isn't the case in ASE either.
	//
	// TDS itself doesn't have a way to explicitly communicate that
	// a cursor has consumed all rows in a data set - and since ASE
	// doesn't implement the one way it could it is assumed that
	// a cursor is finished once a fetch with TDS_CUR_NEXT doesn't
	// return a row.
	//
	// So - instead of relying on information from ASE this boolean is
	// set when more rows are queued in the channel. If no more rows are
	// received this function returns ErrCurNoMoreRows, signaling the
	// cursor finished the result set.
	readMoreRows := false

	_, err := rows.cursor.conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) (bool, error) {
		switch typed := pkg.(type) {
		case *tds.RowPackage:
			rows.rows <- typed
			readMoreRows = true
			return false, nil
		case *tds.RowFmtPackage:
			// TODO: should next return io.EOF if the result set is
			// finished?
			rows.cursor.rowFmt = typed
			return false, nil
		case *tds.OrderByPackage:
			return false, nil
		case *tds.CurInfoPackage:
			// When the result set is exhausted the TDS server
			// deallocates the cursor and notifies the client using
			// tow CurInfoPackages with TDS_CUR_ISTAT_CLOSED and
			// TDS_CUR_ISTAT_DEALLOC.
			if typed.Command != tds.TDS_CUR_CMD_INFORM {
				return true, fmt.Errorf("go-ase: received %T with command %s instead of TDS_CUR_CMD_INFORM",
					typed, typed.Command)
			}

			if typed.Status&tds.TDS_CUR_ISTAT_CLOSED == tds.TDS_CUR_ISTAT_CLOSED {
				// Mark cursor as closed
				rows.cursor.closed = true
				return false, nil
			}

			if typed.Status&tds.TDS_CUR_ISTAT_DEALLOC == tds.TDS_CUR_ISTAT_DEALLOC {
				return true, nil
			}

			return false, nil
		case *tds.DonePackage:
			if typed.Status&tds.TDS_DONE_COUNT == tds.TDS_DONE_COUNT {
				rows.totalRows += int(typed.Count)
				return false, nil
			}

			ok, err := handleDonePackage(typed)
			if err != nil {
				return true, err
			}

			return ok, nil
		case *tds.ReturnStatusPackage:
			if typed.ReturnValue != 0 {
				return true, fmt.Errorf("query failed with return status %d", typed.ReturnValue)
			}
			return false, nil
		default:
			return true, fmt.Errorf("unhandled package type %T: %v", pkg, pkg)
		}
	})

	if readMoreRows {
		return nil
	}

	if err != nil {
		if !rows.rowsClosed {
			close(rows.rows)
			rows.rowsClosed = true
		}
		if errors.Is(err, io.EOF) {
			return ErrCurNoMoreRows
		}
		return fmt.Errorf("error reading next row package: %w", err)
	}

	return nil
}

// Delete deletes the last read row.
func (rows *CursorRows) Delete(ctx context.Context) error {
	delPkg := new(tds.CurDeletePackage)
	delPkg.CursorID = rows.cursor.cursorID
	if rows.cursor.paramFmt != nil {
		delPkg.TableName = rows.cursor.paramFmt.Fmts[0].Table()
	} else if rows.cursor.rowFmt != nil {
		delPkg.TableName = rows.cursor.rowFmt.Fmts[0].Table()
	} else {
		return fmt.Errorf("go-ase: cursor has neither paramFmt nor rowFmt set")
	}

	if err := rows.cursor.conn.Channel.QueuePackage(ctx, delPkg); err != nil {
		return fmt.Errorf("go-ase: error queueing CurDeletePackage: %w", err)
	}

	keyPkg := new(tds.KeyPackage)
	keyPkg.DataType = asetypes.INTN
	keyPkg.Value = int64(rows.readRows - 1)

	if err := rows.cursor.conn.Channel.SendPackage(ctx, keyPkg); err != nil {
		return fmt.Errorf("go-ase: error sending KeyPackage: %w", err)
	}

	if err := finalize(ctx, rows.cursor.conn.Channel); err != nil {
		return err
	}

	return nil
}

// ColumnTypeLength implements the driver.RowsColumnTypeLength interface.
func (rows CursorRows) ColumnTypeLength(index int) (int64, bool) {
	return rows.cursor.rowFmt.Fmts[index].MaxLength(), true
}

func (rows CursorRows) ColumnTypeDatabaseTypeName(index int) string {
	return string(rows.cursor.rowFmt.Fmts[index].DataType())
}
