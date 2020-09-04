// SPDX-FileCopyrightText: 2020 SAP SE
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

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
	_ "github.com/SAP/go-ase/purego"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("eedexample: %v", err)
	}
}

func DoMain() error {
	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	if _, err := db.Exec("sp_adduser nologin"); err != nil {
		var eedError *tds.EEDError
		if errors.As(err, &eedError) {
			fmt.Println("Messages from ASE server:")
			for _, eed := range eedError.EEDPackages {
				fmt.Printf("    %d: %s\n", eed.MsgNumber, eed.Msg)
			}
		}
		return err
	}

	return nil
}
