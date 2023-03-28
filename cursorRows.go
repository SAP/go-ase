// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/SAP/go-dblib/tds"
)

var (
	_ driver.Rows                           = (*CursorRows)(nil)
	_ driver.RowsColumnTypeLength           = (*CursorRows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*CursorRows)(nil)
	_ driver.RowsColumnTypePrecisionScale   = (*CursorRows)(nil)
	_ driver.RowsColumnTypeNullable         = (*CursorRows)(nil)
	_ driver.RowsColumnTypeScanType         = (*CursorRows)(nil)

	ErrCurNoMoreRows = errors.New("no more rows in cursor")
)

// CursorRows is used to iterate over the result set of a cursor.
type CursorRows struct {
	cursor *Cursor
	rows   chan *tds.RowPackage

	baseRows

	// The count of read rows is required to update/delete a row through
	// the cursor.
	readRows  int
	totalRows int
}

// NewCursorRows returns CursorRows for a Cursor.
//
// It does not immediately fetch a result set from the remote. See .Fetch.
func (cursor *Cursor) NewCursorRows() (*CursorRows, error) {
	cursorRows := new(CursorRows)

	cursorRows.cursor = cursor
	cursorRows.rowFmt = func() *tds.RowFmtPackage { return cursor.rowFmt }

	cursorRows.rows = make(chan *tds.RowPackage, cursor.conn.Info.CursorCacheRows)

	return cursorRows, nil
}

// Close closes CursorRows and its associated Cursor.
func (rows *CursorRows) Close() error {
	return rows.cursor.Close(context.Background())
}

// Next implements driver.Rows.
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

// nextPkg is a wrapper to handle reading from the rows channel and
// fetching new rows as needed.
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

// fetch retrieves the next part of the result set from the ASE server.
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
		if !rows.isClosed() {
			close(rows.rows)
			rows.closed = true
		}
		if errors.Is(err, io.EOF) {
			return ErrCurNoMoreRows
		}
		return fmt.Errorf("error reading next row package: %w", err)
	}

	return nil
}
