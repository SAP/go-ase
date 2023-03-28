// SPDX-FileCopyrightText: 2020 SAP SE
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
	"reflect"

	"github.com/SAP/go-dblib/tds"
)

// baseRows is used to share common code between Rows and CursorRows.
type baseRows struct {
	// rowFmt is a workaround to supply the current RowFmtPackage
	// dynamically, as it may change and CursorRows only has access to
	// it through the Cursor.
	rowFmt func() *tds.RowFmtPackage

	// closed effectively disables a rows object if a valid rows object
	// is required but won't receive any data.
	//
	// An example for this are the Direct-/GenericExec methods, which
	// return both a valid Rows and a valid Result.
	// In the case of update statements only the Rows would not receive
	// any data and hence must be prevented from attempting to consume
	// packages.
	//
	// Another example is for rows resulting in a cursor to close the
	// underlying channel used to buffer fields and prevent consuming
	// more packages.
	closed bool
}

func (rows baseRows) isClosed() bool {
	// This method solely exists to silence the check from
	// exhaustivestruct, which reports the attribute 'closed' as unused
	// despite it being used in structs embedding baseRows.
	return rows.closed
}

// fmts returns the list of tds.FieldFmts in the current result set.
//
// FieldFmts marked as hidden are not returned.
func (rows baseRows) fmts() []tds.FieldFmt {
	if rows.rowFmt() == nil {
		return nil
	}

	fmts := []tds.FieldFmt{}

	for _, fieldFmt := range rows.rowFmt().Fmts {
		// ignore hidden columns
		if fmtStatus := tds.RowFmtStatus(fieldFmt.Status()); fmtStatus&tds.TDS_ROW_HIDDEN == tds.TDS_ROW_HIDDEN {
			continue
		}

		fmts = append(fmts, fieldFmt)
	}

	return fmts
}

// Columns implements the driver.Rows interface.
func (rows baseRows) Columns() []string {
	fmts := rows.fmts()
	if len(fmts) == 0 {
		return []string{}
	}

	response := make([]string, len(fmts))

	for i, fieldFmt := range fmts {
		// TODO check if RowFmt is wide and contains column label,
		// catalogue, schema, table
		response[i] = fieldFmt.Name()
	}

	return response
}

// ColumnTypeLength implements the driver.RowsColumnTypeLength interface.
func (rows baseRows) ColumnTypeLength(index int) (int64, bool) {
	if index >= len(rows.fmts()) {
		return 0, false
	}
	return rows.fmts()[index].MaxLength(), true
}

// ColumnTypeDisplayLength returns a best guess of the maximum length
// required to display the values of the column.
func (rows baseRows) ColumnTypeDisplayLength(index int) (int64, bool) {
	if index >= len(rows.fmts()) {
		return 0, false
	}
	return rows.fmts()[index].DisplayMaxLength(), true
}

// ColumnTypeDatabaseTypeName implements the
// driver.RowsColumnTypeDatabaseTypeName interface.
func (rows baseRows) ColumnTypeDatabaseTypeName(index int) string {
	if index >= len(rows.fmts()) {
		return ""
	}
	return rows.fmts()[index].DataType().String()
}

// ColumnTypePrecisionScale implements the
// driver.RowsColumnTypePrecisionScale interface.
func (rows baseRows) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	if index >= len(rows.fmts()) {
		return 0, 0, false
	}

	type PrecisionScaler interface {
		Precision() uint8
		Scale() uint8
	}

	colType, ok := rows.fmts()[index].(PrecisionScaler)
	if !ok {
		return 0, 0, false
	}

	return int64(colType.Precision()), int64(colType.Scale()), true
}

// ColumnTypeNullable implements the
// driver.RowsColumnTypeNullable interface.
func (rows baseRows) ColumnTypeNullable(index int) (bool, bool) {
	if index >= len(rows.fmts()) {
		return false, false
	}

	if rows.fmts()[index].Status()&uint(tds.TDS_ROW_NULLALLOWED) != uint(tds.TDS_ROW_NULLALLOWED) {
		return false, true
	}

	return true, true
}

// ColumnTypeScanType implements the
// driver.RowsColumnTypeScanType interface.
func (rows baseRows) ColumnTypeScanType(index int) reflect.Type {
	if index >= len(rows.fmts()) {
		return nil
	}
	return rows.fmts()[index].DataType().GoReflectType()
}

// Interface satisfaction checks.
var (
	_ driver.Rows                           = (*Rows)(nil)
	_ driver.RowsNextResultSet              = (*Rows)(nil)
	_ driver.RowsColumnTypeLength           = (*Rows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*Rows)(nil)
	_ driver.RowsColumnTypePrecisionScale   = (*Rows)(nil)
	_ driver.RowsColumnTypeNullable         = (*Rows)(nil)
	_ driver.RowsColumnTypeScanType         = (*Rows)(nil)
)

// Rows implements the driver.Rows interface.
type Rows struct {
	Conn *Conn

	baseRows
	RowFmt *tds.RowFmtPackage

	hasNextResultSet bool
}

func (conn *Conn) NewRows() *Rows {
	rows := new(Rows)

	rows.Conn = conn
	rows.rowFmt = func() *tds.RowFmtPackage { return rows.RowFmt }

	return rows
}

// Close implements the driver.Rows interface.
func (rows *Rows) Close() error {
	if rows.isClosed() {
		return nil
	}

	defer func() {
		rows.closed = true
	}()

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

// Next implements the driver.Rows interface.
func (rows *Rows) Next(dst []driver.Value) error {
	if rows.isClosed() || (rows.rowFmt() == nil && len(dst) == 0) {
		return io.EOF
	}

	_, err := rows.Conn.Channel.NextPackageUntil(context.Background(), true,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.RowPackage:
				if len(dst) != len(typed.DataFields) {
					return true, fmt.Errorf("go-ase: received invalid number of destinations, expecting %d destinations, got %d", len(typed.DataFields), len(dst))
				}
				for i := range typed.DataFields {
					dst[i] = typed.DataFields[i].Value()
				}
				return true, nil
			case *tds.RowFmtPackage:
				rows.RowFmt = typed
				rows.hasNextResultSet = true
				return false, io.EOF
			case *tds.OrderByPackage, *tds.OrderBy2Package:
				return false, nil
			case *tds.DonePackage:
				ok, err := handleDonePackage(typed)
				if err != nil {
					return true, fmt.Errorf("go-ase: %w", err)
				}

				return ok, nil
			case *tds.ReturnStatusPackage:
				if typed.ReturnValue != 0 {
					return true, fmt.Errorf("go-ase: query failed with return status %d", typed.ReturnValue)
				}
				return false, nil
			default:
				return true, fmt.Errorf("unhandled package type %T: %v", pkg, pkg)
			}
		},
	)

	if err != nil {
		// database/sql expects only an io.EOF - it doesn't check with
		// errors.Is.
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("go-ase: error reading next row package: %w", err)
	}

	return nil
}

// HasNextResultSet implements the driver.RowsNextResultSet interface.
func (rows *Rows) HasNextResultSet() bool {
	if !rows.hasNextResultSet {
		return false
	}
	rows.hasNextResultSet = false
	return true
}

// NextResultSet implements the driver.RowsNextResultSet interface.
func (rows *Rows) NextResultSet() error {
	if rows.isClosed() {
		return io.EOF
	}

	// discard all RowPackage until either end of communication or next
	// RowFmtPackage
	_, err := rows.Conn.Channel.NextPackageUntil(context.Background(), false,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.RowFmtPackage:
				rows.RowFmt = typed
				rows.hasNextResultSet = true
				return false, nil
			case *tds.RowPackage, *tds.OrderByPackage, *tds.OrderBy2Package:
				return true, nil
			case *tds.DonePackage:
				if typed.Status&tds.TDS_DONE_MORE == tds.TDS_DONE_MORE {
					return false, nil
				}
				return true, fmt.Errorf("go-ase: no next result set: %w", io.EOF)
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
