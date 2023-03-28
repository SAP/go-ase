// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows the use of transactions using the database/sql
// interface and the pure go driver.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-ase/examples"
	"github.com/SAP/go-dblib/dsn"
)

const (
	exampleName  = "transaction"
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

	return Test(db)
}

func Test(db *sql.DB) error {
	fmt.Println("opening transaction")
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}

	fmt.Println("inserting values")
	if _, err := tx.Exec("insert into "+tableName+" (a, b) values (?, ?)", math.MaxInt32, "a string"); err != nil {
		return fmt.Errorf("failed to insert values: %w", err)
	}

	fmt.Println("preparing statement")
	stmt, err := tx.Prepare("select * from " + tableName + " where a=?")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("error closing statement: %v", err)
		}
	}()

	fmt.Println("executing prepared statement")
	rows, err := stmt.Query(math.MaxInt32)
	if err != nil {
		return fmt.Errorf("error querying with prepared statement: %w", err)
	}

	var a int
	var b string
	for rows.Next() {
		if err := rows.Scan(&a, &b); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		fmt.Printf("a: %d\n", a)
		fmt.Printf("b: %s\n", b)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows reported error: %w", err)
	}

	fmt.Println("committing transaction")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
