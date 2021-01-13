// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to utilize the GenericExecer interface if both
// driver.Rows and driver.Result of a statement are required.
//
// This has the drawback that it essentially circumvents the
// database/sql interface for most interactions - the only advantage
// over using the driver directly will be the use of connection pooling
// of database/sql.
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"

	_ "github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/dsn"
)

type GenericExecer interface {
	GenericExec(context.Context, string, []driver.NamedValue) (driver.Rows, driver.Result, error)
}

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("genericexec example: %v", err)
	}
}

func DoMain() error {
	dsn, err := dsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("genericexec example: error closing db: %v", err)
		}
	}()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error getting conn: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("genericexec example: error closing conn: %v", err)
		}
	}()

	return conn.Raw(func(driverConn interface{}) error {
		if err := rawProcess(driverConn); err != nil {
			return fmt.Errorf("error in rawProcess: %w", err)
		}
		return nil
	})
}

func rawProcess(driverConn interface{}) error {
	execer, ok := driverConn.(GenericExecer)
	if !ok {
		return errors.New("invalid driver, must support GenericExecer")
	}

	rows, _, err := execer.GenericExec(context.Background(), "select @@version", nil)
	if err != nil {
		return fmt.Errorf("error in genericexec: %w", err)
	}
	defer rows.Close()

	args := []driver.Value{""}

	for {
		if err := rows.Next(args); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error scanning row: %w", err)
		}

		if args[0] == "" {
			return fmt.Errorf("version is empty after scanning")
		}

		// The output can't contain the version itself for the example
		// test.
		fmt.Println("version was read")
		log.Printf("genericexec example: %s", args[0])
	}

	return nil
}
