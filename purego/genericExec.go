// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/tds"
)

func (c *Conn) GenericExec(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
	if len(args) == 0 {
		rows, result, err := c.language(ctx, query)
		if err != nil {
			return nil, nil, fmt.Errorf("go-ase: error executing statement: %w", err)
		}
		return rows, result, nil
	}

	stmt, err := c.NewStmt(ctx, "", query, true)
	if err != nil {
		return nil, nil, fmt.Errorf("go-ase: error preparing dynamic SQL: %w", err)
	}

	for i := range args {
		if err := stmt.CheckNamedValue(&args[i]); err != nil {
			return nil, nil, fmt.Errorf("go-ase: error checking argument: %w", err)
		}
	}

	rows, result, err := stmt.exec(ctx, args)
	if err != nil {
		return nil, nil, fmt.Errorf("go-ase: error executing dynamic SQL: %w", err)
	}

	return rows, result, nil
}

func (c *Conn) genericResults(ctx context.Context) (driver.Rows, driver.Result, error) {
	rows := &Rows{Conn: c}
	result := &Result{}

	_, err := c.Channel.NextPackageUntil(ctx, true,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.RowFmtPackage:
				rows.RowFmt = typed
				return true, nil
			case *tds.DonePackage:
				if typed.Status&tds.TDS_DONE_COUNT == tds.TDS_DONE_COUNT {
					result.rowsAffected = int64(typed.Count)
				}

				if typed.Status&tds.TDS_DONE_MORE == tds.TDS_DONE_MORE {
					return false, nil
				}

				if typed.Status&tds.TDS_DONE_PROC == tds.TDS_DONE_PROC || typed.Status&tds.TDS_DONE_FINAL == tds.TDS_DONE_FINAL {
					return true, nil
				}

				return false, fmt.Errorf("%T does not have status TDS_DONE_COUNT or TDS_DONE_FINAL set: %s",
					typed, typed)
			case *tds.ReturnStatusPackage:
				if typed.ReturnValue != 0 {
					return false, fmt.Errorf("received return status %d", typed.ReturnValue)
				}
				return false, nil
			default:
				return false, fmt.Errorf("unhandled package type %T", typed)
			}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	if rows.RowFmt == nil {
		return nil, result, nil
	}

	return rows, result, nil
}
