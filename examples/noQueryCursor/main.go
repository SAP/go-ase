// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// go-ase uses cursor by default for all SQL queries.
//
// This is mostly useful for queries with larger result sets where
// receiving all results immediately would either hurt performance or
// use up too many resources.
//
// If only queries with small result sets are used it may be viable to
// disable using cursors by setting the .NoQueryCursor attribute on the
// DSN info to true.
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
	exampleName  = "orderBy"
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

	dropTable, err := examples.CreateDropTable(db, tableName, "a bigint, b varchar(30)")
	if err != nil {
		return err
	}
	defer dropTable()

	type Value struct {
		a int
		b string
	}

	values := []Value{
		{1, "one"},
		{0, "zero"},
		{4, "four"},
		{3, "three"},
	}

	for _, value := range values {
		if _, err := db.Exec("insert into "+tableName+" values (?, ?)", value.a, value.b); err != nil {
			return fmt.Errorf("failed to insert values %q: %w", value, err)
		}
	}

	fmt.Println("Query without cursor")
	info.NoQueryCursor = true
	if err := query(db); err != nil {
		return err
	}

	fmt.Println("Query with cursor")
	info.NoQueryCursor = false
	if err := query(db); err != nil {
		return err
	}

	return nil
}

func query(db *sql.DB) error {
	fmt.Println("Querying values from table without ordering")
	rows, err := db.Query("select * from " + tableName)
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	if err := display(rows); err != nil {
		return err
	}

	fmt.Println("Querying values from table with ordering")
	rows, err = db.Query("select * from " + tableName + " order by a")
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	if err := display(rows); err != nil {
		return err
	}

	return nil
}

func display(rows *sql.Rows) error {
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
