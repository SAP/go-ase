// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/SAP/go-ase/libase/tds"
	"github.com/SAP/go-ase/libase/term"
	ase "github.com/SAP/go-ase/purego"
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatalf("goase failed: %v", err)
	}
}

func doMain() error {
	dsn, err := term.Dsn()
	if err != nil {
		return fmt.Errorf("error parsing DSN from env: %w", err)
	}

	connector, err := ase.NewConnectorWithHooks(dsn, updateDatabaseName)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	if len(flag.Args()) > 0 {
		// Positional arguments were supplied, execute these as SQL
		// statements
		query := strings.Join(flag.Args(), " ") + ";"
		return term.ParseAndExecQueries(db, query)
	}

	return term.Repl(db)
}

func updateDatabaseName(typ tds.EnvChangeType, oldValue, newValue string) {
	if typ != tds.TDS_ENV_DB {
		return
	}

	term.PromptDatabaseName = newValue
}
