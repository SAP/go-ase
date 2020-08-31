// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql/driver"
	"fmt"
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
		err := stmt.CheckNamedValue(&args[i])
		if err != nil {
			return nil, nil, fmt.Errorf("go-ase: error checking argument: %w", err)
		}
	}

	rows, result, err := stmt.exec(ctx, args)
	if err != nil {
		return nil, nil, fmt.Errorf("go-ase: error executing dynamic SQL: %w", err)
	}

	return rows, result, nil
}
