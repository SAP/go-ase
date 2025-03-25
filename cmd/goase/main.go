// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/tds"
	"github.com/SAP/go-dblib/term"

	"github.com/spf13/pflag"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatalf("goase failed: %v", err)
	}
}

func doMain() error {
	info, flagset, err := ase.NewInfoWithFlags()
	if err != nil {
		return fmt.Errorf("error creating info: %w", err)
	}

	// Use pflag to merge flagsets
	flags := pflag.NewFlagSet("goase", pflag.ContinueOnError)

	// Merge info flagset
	flags.AddGoFlagSet(flagset)

	// Merge stdlib flag arguments
	flags.AddGoFlagSet(flag.CommandLine)

	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if err := dsn.FromEnv("ASE", info); err != nil {
		return fmt.Errorf("error reading values from environment: %w", err)
	}

	connector, err := ase.NewConnectorWithHooks(info,
		[]tds.EnvChangeHook{updateDatabaseName},
		[]tds.EEDHook{logEED},
	)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	return term.Entrypoint(db, flags.Args())
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
