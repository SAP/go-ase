// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to retrieve the wrapped tds.EEDError to access
// messages sent by the TDS server in the context of an SQL statement.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/tds"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("eedexample: %v", err)
	}
}

func DoMain() error {
	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	fmt.Println("sp_adduser")
	if _, err := db.Exec("sp_adduser nologin"); err != nil {
		var eedError *tds.EEDError
		if errors.As(err, &eedError) {
			fmt.Println("Messages from ASE server:")
			for _, eed := range eedError.EEDPackages {
				fmt.Printf("    %d: %s\n", eed.MsgNumber, eed.Msg)
			}
		}
	}

	fmt.Println("create table eed_example_tab")
	if _, err := db.Exec("create table eed_example_tab values (int, string)"); err != nil {
		var eedError *tds.EEDError
		if errors.As(err, &eedError) {
			fmt.Println("Messages from ASE server:")
			for _, eed := range eedError.EEDPackages {
				fmt.Printf("    %d: %s\n", eed.MsgNumber, eed.Msg)
			}
		}
	}

	fmt.Println("create database")
	if _, err := db.Exec("create database"); err != nil {
		var eedError *tds.EEDError
		if errors.As(err, &eedError) {
			fmt.Println("Messages from ASE server:")
			for _, eed := range eedError.EEDPackages {
				fmt.Printf("    %d: %s\n", eed.MsgNumber, eed.Msg)
			}
		}
	}

	return nil
}
