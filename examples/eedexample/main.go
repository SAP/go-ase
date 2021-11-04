// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to retrieve the wrapped tds.EEDError to access
// messages sent by the TDS server in the context of an SQL statement.
//
// This can be used to introspect errors and log the specific error
// messages sent by the ASE server.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-ase/examples"
	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/tds"
)

const (
	exampleName  = "eedExample"
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

	return Test(db)
}

func Test(db *sql.DB) error {

	// `string` is not a valid column type in ASE, causing the server to
	// return an error in an EEDPackage
	if _, err := db.Exec("create table " + tableName + " (a int, b string)"); err != nil {
		// tds.EEDError is a wrapper to make multiple EEDPackages
		// available
		var eedError *tds.EEDError
		if errors.As(err, &eedError) {
			fmt.Println("Messages from ASE server:")
			for _, eed := range eedError.EEDPackages {
				// eed.MsgNumber is the error number and can be
				// referenced in the ASE documentation.
				// eed.Msg contains the verbatim error message as sent
				// by the ASE server.
				fmt.Printf("  %d: %s\n", eed.MsgNumber, eed.Msg)
			}
		} else {
			fmt.Printf("Unexpected error received: %v", err)
		}
	}

	return nil
}
