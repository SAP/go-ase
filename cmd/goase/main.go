package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/SAP/go-ase/libase/term"
	_ "github.com/SAP/go-ase/purego"
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatalf("goase failed: %v", err)
	}
}

func doMain() error {
	db, err := sql.Open("ase", term.Dsn().AsSimple())
	if err != nil {
		return fmt.Errorf("goase: failed to connect to database: %w", err)
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
