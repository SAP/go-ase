// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package libtest

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
)

// SetupDB creates a database and sets .Database on the passed testDsn
func SetupDB(testDsn *libdsn.Info) error {
	db, err := sql.Open("ase", testDsn.AsSimple())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(context.Background(), "use master")
	if err != nil {
		return fmt.Errorf("failed to switch context to master: %w", err)
	}

	testDatabase := "test" + RandomNumber()

	_, err = conn.ExecContext(context.Background(), fmt.Sprintf("if db_id('%s') is not null drop database %s", testDatabase, testDatabase))
	if err != nil {
		return fmt.Errorf("error on conditional drop of database: %w", err)
	}

	_, err = conn.ExecContext(context.Background(), "create database "+testDatabase)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	_, err = conn.ExecContext(context.Background(), "use "+testDatabase)
	if err != nil {
		return fmt.Errorf("failed to switch context to %s: %w", testDatabase, err)
	}

	testDsn.Database = testDatabase
	return nil
}

// TeardownDB deletes the database indicated by .Database of the passed
// testDsn and unsets the member.
func TeardownDB(testDsn *libdsn.Info) error {
	db, err := sql.Open("ase", testDsn.AsSimple())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(context.Background(), "use master")
	if err != nil {
		return fmt.Errorf("failed to switch context to master: %w", err)
	}

	_, err = conn.ExecContext(context.Background(), "drop database "+testDsn.Database)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	testDsn.Database = ""
	return nil
}

// SetupTableInsert creates a table with the passed type and inserts all
// passed samples as rows.
func SetupTableInsert(db *sql.DB, tableName, aseType string, samples ...interface{}) (*sql.Rows, func() error, error) {
	_, err := db.Exec(fmt.Sprintf("create table %s (a %s)", tableName, aseType))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create table: %w", err)
	}

	stmt, err := db.Prepare(fmt.Sprintf("insert into %s (a) values (?)", tableName))
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, sample := range samples {
		_, err := stmt.Exec(sample)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute prepared statement with %v: %w", sample, err)
		}
	}

	rows, err := db.Query("select * from " + tableName)
	if err != nil {
		return nil, nil, fmt.Errorf("error selecting from %s: %w", tableName, err)
	}

	teardownFn := func() error {
		_, err := db.Exec("drop table " + tableName)
		return err
	}

	return rows, teardownFn, nil
}
