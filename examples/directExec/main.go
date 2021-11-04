// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to utilize the DirectExecer interface to
// receive both the driver.Rows and driver.Result of a SQL query.
//
// go-ase offers two methods on go-ase connections - GenericExec and
// DirectExec. They are functionally identical as DirectExec is
// a wrapper around GenericExec and takes interface{}s als values for
// the placeholders within an SQL statement, rather than driver.Values.
//
// E.g.:
//   if rows, result, err := conn.GenericExec(ctx, "select * from table where a = ?", driver.Value{5}); err != nil {
//     return err
//   }
//
//   if rows, result, err := conn.DirectExec(ctx, "select * from table where a = ?", 5); err != nil {
//     return err
//   }
//
// These methods are primarily useful when using stored procedures which
// may return both rows and the number of affected rows.
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-ase/examples"
	"github.com/SAP/go-dblib/dsn"
)

const (
	exampleName  = "directExec"
	databaseName = exampleName + "DB"
	tableName    = databaseName + ".." + exampleName + "Table"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("%s failed: %v", exampleName, err)
	}
}

func DoMain() error {
	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	dropDB, err := examples.CreateDropDatabase(info, databaseName)
	if err != nil {
		return err
	}
	defer dropDB()

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer db.Close()

	dropTable, err := examples.CreateDropTable(db, tableName, "a int")
	if err != nil {
		return err
	}
	defer dropTable()

	if _, err := db.Exec("insert into "+tableName+" values (?)", 1); err != nil {
		return err
	}

	return Test(db)
}

func Test(db *sql.DB) error {
	// To access the .DirectExec method the underlying go-ase.Conn must
	// be used.
	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error getting conn: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing conn: %v", err)
		}
	}()

	// Retrieve the underlying go-ase.Conn
	return conn.Raw(func(driverConn interface{}) error {
		if err := rawProcess(driverConn); err != nil {
			return fmt.Errorf("error in rawProcess: %w", err)
		}
		return nil
	})
}

type DirectExecer interface {
	DirectExec(context.Context, string, ...interface{}) (driver.Rows, driver.Result, error)
}

func rawProcess(driverConn interface{}) error {
	execer, ok := driverConn.(DirectExecer)
	if !ok {
		return fmt.Errorf("invalid driver connection %T, must support DirectExecer interface", driverConn)
	}

	rows, result, err := execer.DirectExec(context.Background(), "update "+tableName+" set a = ? where a = 1", 5)
	if err != nil {
		return fmt.Errorf("error executing SQL statement: %w", err)
	}
	defer rows.Close()

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows count: %w", err)
	}

	fmt.Printf("affected rows: %d\n", rowsAffected)

	args := []driver.Value{0}

	for {
		// The returned rows implements the database/sql/driver.Rows
		// interface, which is slightly different from the
		// database/sql.Rows signature.
		if err := rows.Next(args); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error scanning row: %w", err)
		}

		fmt.Printf("a: %d\n", args[0])
	}

	return nil
}
