// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// TODO describe what the example does or shows here
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
	exampleName  = "TODO"
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

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer db.Close()

	dropDB, err := examples.CreateDropDatabase(db, databaseName)
	if err != nil {
		return err
	}
	defer dropDB()

	dropTable, err := examples.CreateDropTable(db, tableName, "a int")
	if err != nil {
		return err
	}
	defer dropTable()

	return Test(db)
}

func Test(db *sql.DB) error {
	fmt.Println("TODO replace with test code")
	return nil
}
