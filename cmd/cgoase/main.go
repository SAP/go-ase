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

	"github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/term"
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatalf("cgoase failed: %v", err)
	}
}

func doMain() error {
	cgo.GlobalServerMessageBroker.RegisterHandler(handleMessage)
	cgo.GlobalClientMessageBroker.RegisterHandler(handleMessage)

	db, err := sql.Open("ase", term.Dsn().AsSimple())
	if err != nil {
		return fmt.Errorf("cgoase: failed to connect to database: %w", err)
	}
	defer db.Close()

	if len(flag.Args()) > 0 {
		// Positional arguments were supplied, execute these as SQL
		// statements
		query := strings.Join(flag.Args(), " ") + ";"
		return term.ParseAndExecQueries(db, query)
	}

	return term.Repl(db)
}

func handleMessage(msg cgo.Message) {
	if msg.MessageSeverity() == 10 {
		return
	}

	log.Printf("Msg %d: %s", msg.MessageNumber(), msg.Content())
}
