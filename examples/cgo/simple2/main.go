// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows a simple interaction with a TDS server using the
// database/sql interface, prepared statements and the cgo-based driver.
package main

import (
	"database/sql"
	"fmt"
	"log"

	ase "github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/libdsn"
)

func main() {
	err := DoMain()
	if err != nil {
		log.Fatalf("godb failed: %v", err)
	}
}

func DoMain() error {
	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	connector, err := ase.NewConnector(dsn)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("Closing database failed: %v", err)
		}
	}()

	database := "newDB"

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

	table := fmt.Sprintf("%s..testTable", database)

	fmt.Printf("creating table %s\n", table)
	if _, err := db.Exec("create table " + table + " (a tinyint)"); err != nil {
		return fmt.Errorf("error creating table %s: %w", table, err)
	}

	fmt.Printf("inserting values into table %s\n", table)
	if _, err := db.Exec("insert into "+table+" (a) values (?)", 5); err != nil {
		return fmt.Errorf("error inserting value %d into table %s: %w", 5, table, err)
	}

	fmt.Println("preparing statement")
	stmt, err := db.Prepare(fmt.Sprintf("select * from %s where a=?", table))
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(5)
	if err != nil {
		return fmt.Errorf("error querying with prepared statement: %w", err)
	}

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
