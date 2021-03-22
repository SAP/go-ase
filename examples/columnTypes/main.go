// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to retrieve column type information.
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SAP/go-ase"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("columnTypes failed: %v", err)
	}
}

func DoMain() error {
	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	connector, err := ase.NewConnector(info)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	fmt.Println("opening database")
	db := sql.OpenDB(connector)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Closing database failed: %v", err)
		}
	}()

	database := "columnTypesDB"

	fmt.Println("switching to master")
	if _, err := db.Exec("use master"); err != nil {
		return fmt.Errorf("error switching to master: %w", err)
	}

	fmt.Printf("dropping database %s if exists\n", database)
	if _, err := db.Exec(fmt.Sprintf("if db_id('%s') is not null drop database %s", database, database)); err != nil {
		return fmt.Errorf("error dropping database: %w", err)
	}

	fmt.Printf("creating database %s\n", database)
	if _, err := db.Exec("create database " + database); err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}
	defer func() {
		fmt.Println("teardown: switching to master")
		if _, err := db.Exec("use master"); err != nil {
			log.Printf("teardown: error switching to master: %v", err)
			return
		}

		fmt.Printf("teardown: dropping database %s\n", database)
		if _, err := db.Exec("drop database " + database); err != nil {
			log.Printf("teardown: error dropping database %s: %v", database, err)
		}
	}()

	table := fmt.Sprintf("%s..columnTypes_tab", database)

	fmt.Printf("creating table %s\n", table)
	if _, err := db.Exec("create table " + table + " (a bigint, b numeric(32,0), c decimal(16,2) null, d char(8), e varchar(32) null)"); err != nil {
		return fmt.Errorf("error creating table %s: %w", table, err)
	}

	fmt.Printf("inserting values into table %s\n", table)
	if _, err := db.Exec("insert into " + table + "(a, b, c, d, e) values (123456, 123456, 123.45, '', 'test')"); err != nil {
		return fmt.Errorf("error inserting values into table %s: %w", table, err)
	}

	fmt.Println("querying values from table")
	rows, err := db.Query(fmt.Sprintf("select * from %s", table))
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
		fmt.Printf("type: %s\n", col.DatabaseTypeName())
		if length, ok := col.Length(); ok {
			fmt.Printf("length = %d\n", length)
		}
		if precision, scale, ok := col.DecimalSize(); ok {
			fmt.Printf("precision = %d\n", precision)
			fmt.Printf("scale = %d\n", scale)
		}
		if nullable, ok := col.Nullable(); ok {
			fmt.Printf("nullable = %t\n", nullable)
		}
		fmt.Printf("scan type: %v\n", col.ScanType())
	}

	return nil
}
