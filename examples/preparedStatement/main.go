// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows a simple interaction with a TDS server using the
// database/sql interface, prepared statements and the pure go driver.
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-ase/examples"
	"github.com/SAP/go-dblib/dsn"
)

const (
	exampleName  = "preparedStatement"
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

	return Test(db)
}

func Test(db *sql.DB) error {
	fmt.Printf("inserting values into table %s\n", tableName)
	if _, err := db.Exec("insert into "+tableName+" (a) values (?)", 5); err != nil {
		return fmt.Errorf("error inserting value %d into table %s: %w", 5, tableName, err)
	}

	// go-ase uses prepared statements with cursors by default, using
	// prepared statements explicitly only nets the benefit of
	// using the sql.Stmt struct and improved perforamnce if a statement
	// must be executed multiple times with different parameters.
	fmt.Println("preparing statement")
	stmt, err := db.Prepare("select * from " + tableName + " where a=?")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	fmt.Println("executing prepared statement")
	rows, err := stmt.Query(5)
	if err != nil {
		return fmt.Errorf("error querying with prepared statement: %w", err)
	}
	defer rows.Close()

	var a int
	for rows.Next() {
		if err := rows.Scan(&a); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		fmt.Printf("a: %d\n", a)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows reported error: %w", err)
	}

	return nil
}
