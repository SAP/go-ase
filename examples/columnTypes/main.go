// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to retrieve column type information.
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
	exampleName  = "columnTypes"
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

	dropTable, err := examples.CreateDropTable(db, tableName, "a bigint, b numeric(32,0), c decimal(16,2) null, d char(8), e varchar(32) null")
	if err != nil {
		return err
	}
	defer dropTable()

	return Test(db)
}

func Test(db *sql.DB) error {
	// There are no values in the table - but a valid *sql.Rows will
	// still be returned, which can be used to inspect the layout of the
	// table.
	//
	// Even if the table has values using a 'select *' to inspect the
	// table columns won't impact performance as cursors are used by
	// default.
	fmt.Println("querying table")
	rows, err := db.Query(fmt.Sprintf("select * from %s", tableName))
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to retrieve column types: %w\n", err)
	}

	for _, col := range colTypes {
		fmt.Printf("column-name: %s\n", col.Name())
		fmt.Printf("  type: %s\n", col.DatabaseTypeName())
		if length, ok := col.Length(); ok {
			fmt.Printf("  length = %d\n", length)
		}
		if precision, scale, ok := col.DecimalSize(); ok {
			fmt.Printf("  precision = %d\n", precision)
			fmt.Printf("  scale = %d\n", scale)
		}
		if nullable, ok := col.Nullable(); ok {
			fmt.Printf("  nullable = %t\n", nullable)
		}
		fmt.Printf("  scan type: %v\n", col.ScanType())
	}

	return nil
}
