// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/SAP/go-ase/libase/tds"
)

var (
	_ driver.Rows                           = (*Rows)(nil)
	_ driver.RowsNextResultSet              = (*Rows)(nil)
	_ driver.RowsColumnTypeLength           = (*Rows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*Rows)(nil)
)

type Rows struct {
	Conn   *Conn
	RowFmt *tds.RowFmtPackage
}

func (rows Rows) Columns() []string {
	// TODO ignore hidden columns
	response := make([]string, len(rows.RowFmt.Fmts))

	for i, fieldFmt := range rows.RowFmt.Fmts {
		// TODO check if RowFmt is wide and contains column label,
		// catalgoue, schema, table
		response[i] = fieldFmt.Name()
	}

	return response
}

func (rows *Rows) Close() error {
	for {
		if err := rows.NextResultSet(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("go-ase: error consuming result sets: %w", err)
		}
	}

	return nil
}

func (rows *Rows) Next(dst []driver.Value) error {
	returnStatus := -1
	recvErr := false

	_, err := rows.Conn.Channel.NextPackageUntil(context.Background(), true,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.RowPackage:
				for i := range dst {
					dst[i] = typed.DataFields[i].Value()
				}
				return true, nil
			case *tds.RowFmtPackage:
				// TODO: should next return io.EOF if the result set is
				// finished?
				rows.RowFmt = typed
				return false, nil
			case *tds.OrderByPackage:
				return false, nil
			case *tds.DonePackage:
				if typed.Status == tds.TDS_DONE_COUNT {
					return true, io.EOF
				}

				if typed.Status&tds.TDS_DONE_MORE == tds.TDS_DONE_MORE {
					return false, nil
				}

				if typed.Status&tds.TDS_DONE_ERROR == tds.TDS_DONE_ERROR {
					recvErr = true
					return false, nil
				}

				if typed.Status&tds.TDS_DONE_INXACT == tds.TDS_DONE_INXACT {
					return false, nil
				}

				if typed.Status&tds.TDS_DONE_PROC == tds.TDS_DONE_PROC {
					return false, nil
				}

				if typed.Status == tds.TDS_DONE_FINAL {
					if returnStatus > 0 {
						return true, fmt.Errorf("go-ase: query failed with return status %d", returnStatus)
					}
					if recvErr {
						return true, fmt.Errorf("go-ase: query failed with errors")
					}
					return true, io.EOF
				}
				return false, nil
			case *tds.ReturnStatusPackage:
				returnStatus = int(typed.ReturnValue)
				return false, nil
			default:
				return true, fmt.Errorf("unhandled package type %T: %v", pkg, pkg)
			}
		},
	)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("go-ase: error reading next row package: %w", err)
	}

	return nil
}

func (rows *Rows) HasNextResultSet() bool {
	// TODO this doesn't seem good.
	return rows.NextResultSet() != nil
}

func (rows *Rows) NextResultSet() error {
	// discard all RowPackage until either end of communication or next
	// RowFmtPackage
	_, err := rows.Conn.Channel.NextPackageUntil(context.Background(), false,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.RowFmtPackage:
				rows.RowFmt = typed
				return false, nil
			case *tds.RowPackage, *tds.OrderByPackage:
				return true, nil
			case *tds.DonePackage:
				return true, io.EOF
			default:
				return false, fmt.Errorf("unhandled package type %T: %v", pkg, pkg)
			}
		},
	)

	if err != nil {
		if errors.Is(err, tds.ErrNoPackageReady) || errors.Is(err, io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("go-ase: error reading next package: %w", err)
	}

	return nil
}

func (rows Rows) ColumnTypeLength(index int) (int64, bool) {
	return rows.RowFmt.Fmts[index].MaxLength(), true
}

func (rows Rows) ColumnTypeDatabaseTypeName(index int) string {
	return string(rows.RowFmt.Fmts[index].DataType())
}
