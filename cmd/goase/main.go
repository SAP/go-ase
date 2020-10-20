// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/tds"
	"github.com/SAP/go-dblib/term"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatalf("goase failed: %v", err)
	}
}

func doMain() error {
	dsn, err := term.Dsn()
	if err != nil {
		return fmt.Errorf("error parsing DSN from env: %w", err)
	}

	connector, err := ase.NewConnectorWithHooks(dsn,
		[]tds.EnvChangeHook{updateDatabaseName},
		[]tds.EEDHook{logEED},
	)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	return term.Entrypoint(db)
}

func updateDatabaseName(typ tds.EnvChangeType, oldValue, newValue string) {
	if typ != tds.TDS_ENV_DB {
		return
	}

	term.PromptDatabaseName = newValue
}

func logEED(eed tds.EEDPackage) {
	fmt.Println(eed.Msg)
}
