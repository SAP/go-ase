// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/integration"
)

func CreateDropDatabase(info *ase.Info, databaseName string) (func() error, error) {
	if err := integration.SetupDB(context.Background(), info, databaseName); err != nil {
		return nil, err
	}

	return func() error {
		return integration.TeardownDB(context.Background(), info)
	}, nil
}

func CreateDropTable(db *sql.DB, tableName, layout string) (func() error, error) {
	if _, err := db.Exec(fmt.Sprintf("create table %s (%s)", tableName, layout)); err != nil {
		return nil, fmt.Errorf("error creating table %q: %w", tableName, err)
	}

	fn := func() error {
		if _, err := db.Exec(fmt.Sprintf("if object_id('%s') is not null drop table %s", tableName, tableName)); err != nil {
			return fmt.Errorf("error dropping table %q: %w", tableName, err)
		}

		return nil
	}

	return fn, nil
}
