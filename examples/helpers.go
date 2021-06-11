// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SAP/go-dblib/integration"
)

func CreateDropDatabase(db *sql.DB, databaseName string) (func() error, error) {
	if _, err := db.Exec("use master"); err != nil {
		return nil, fmt.Errorf("error switching to master: %w", err)
	}

	integration.DBCreateLock.Lock()
	defer integration.DBCreateLock.Unlock()

	if _, err := db.Exec("create database " + databaseName); err != nil {
		return nil, fmt.Errorf("error creating database %q: %w", databaseName, err)
	}

	fn := func() error {
		conn, err := db.Conn(context.Background())
		if err != nil {
			return fmt.Errorf("teardown %q: error getting db.Conn: %w", databaseName, err)
		}
		defer conn.Close()

		if _, err := conn.ExecContext(context.Background(), "use master"); err != nil {
			return fmt.Errorf("teardown %q: error switching to master: %w", databaseName, err)
		}

		if _, err := conn.ExecContext(context.Background(), fmt.Sprintf("if db_id('%s') is not null drop database %s", databaseName, databaseName)); err != nil {
			return fmt.Errorf("teardown %q: error dropping database %q: %w", databaseName, databaseName, err)
		}

		return nil
	}

	return fn, nil
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
