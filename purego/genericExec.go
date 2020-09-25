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

	"github.com/SAP/go-ase/libase"
	"github.com/SAP/go-ase/libase/tds"
)

// DirectExec is a wrapper for GenericExec and meant to be used when
// directly accessing this library, rather than using database/sql.
//
// The primary advantage are the variadic args, which can be normal
// values and are automatically transformed to driver.NamedValues for
// GenericExec.
func (c *Conn) DirectExec(ctx context.Context, query string, args ...interface{}) (driver.Rows, driver.Result, error) {
	var namedArgs []driver.NamedValue
	if len(args) > 0 {
		values := make([]driver.Value, len(args))
		for i, arg := range args {
			values[i] = driver.Value(arg)
		}
		namedArgs = libase.ValuesToNamedValues(values)
	}
	return c.GenericExec(ctx, query, namedArgs)
}

// GenericExec is the central method through which SQL statements are
// sent to ASE.
func (c *Conn) GenericExec(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
	if len(args) == 0 {
		rows, result, err := c.language(ctx, query)
		if err != nil && !errors.Is(err, io.EOF) {
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

	rows, result, err := stmt.GenericExec(ctx, args)
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
				return false, fmt.Errorf("go-ase: unhandled package type %T", typed)
			}
		},
	)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, nil, err
	}

	if rows.RowFmt == nil {
		return nil, result, nil
	}

	return rows, result, nil
}
