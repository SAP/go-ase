// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows a simple interaction with a TDS server using the
// database/sql interface and the pure go driver.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"

	_ "github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/dsn"
)

func main() {
	if err := DoMain(); err != nil {
		log.Printf("Failed: %v", err)
		os.Exit(1)
	}
}

func DoMain() error {
	dsn, err := dsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	fmt.Println("Opening database")
	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer db.Close()

	if _, err = db.Exec("if object_id('simple_tab') is not null drop table simple_tab"); err != nil {
		return fmt.Errorf("failed to drop table 'simple_tab': %w", err)
	}

	fmt.Println("Creating table 'simple_tab'")
	if _, err = db.Exec("create table simple_tab (a int, b char(30))"); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Printf("Writing a=%d, b='a string' to table\n", math.MaxInt32)
	if _, err = db.Exec("insert into simple_tab (a, b) values (?, ?)", math.MaxInt32, "a string"); err != nil {
		return fmt.Errorf("failed to insert values: %w", err)
	}

	fmt.Println("Querying values from table")
	rows, err := db.Query("select * from simple_tab")
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	fmt.Println("Displaying results of query")
	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to retrieve column names: %w", err)
	}

	fmt.Printf("| %-10s | %-30s |\n", colNames[0], colNames[1])
	format := "| %-10d | %-30s |\n"

	var a int
	var b string

	for rows.Next() {
		if err = rows.Scan(&a, &b); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Printf(format, a, b)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	return nil
}
