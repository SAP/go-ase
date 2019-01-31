package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase"
	"github.com/bgentry/speakeasy"
)

var (
	fHost = flag.String("H", "localhost", "database hostname")
	fPort = flag.String("P", "4901", "database sql port")
	fUser = flag.String("u", "sa", "database user name")
	fPass = flag.String("p", "", "database user password")
)

func exec(db *sql.DB, q string) error {
	result, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("Executing the statement failed: %v", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Retrieving the affected rows failed: %v", err)
	}

	fmt.Printf("Rows affected: %d\n", affectedRows)
	return nil
}

func query(db *sql.DB, q string) error {
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("Query failed: %v", err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("Failed to retrieve column names: %v", err)
	}

	fmt.Printf("|")
	for _, colName := range colNames {
		fmt.Printf(" %s |", colName)
	}
	fmt.Printf("\n")

	cells := make([]interface{}, len(colNames))

	cellsRef := make([]interface{}, len(colNames))
	for i := range cells {
		cellsRef[i] = &(cells[i])
	}

	for rows.Next() {
		err := rows.Scan(cellsRef...)
		if err != nil {
			return fmt.Errorf("Error retrieving rows: %v", err)
		}

		for _, cell := range cells {
			fmt.Printf("| %v ", cell)
		}
		fmt.Printf("|\n")
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("Error preparing rows: %v", err)
	}

	return nil
}

func subcmd(db *sql.DB, part string) {
	partS := strings.Split(part, " ")
	cmd := partS[0]
	q := strings.Join(partS[1:], " ")

	switch cmd {
	case "exec":
		err := exec(db, q)
		if err != nil {
			log.Printf("Exec errored: %v", err)
		}
	case "query":
		err := query(db, q)
		if err != nil {
			log.Printf("Query errored: %v", err)
		}
	default:
		log.Printf("Unknown command: %s", cmd)
	}

}

func main() {
	flag.Parse()
	pass := *fPass
	var err error
	if len(pass) == 0 {
		pass, err = speakeasy.Ask("Please enter the password of user " + *fUser + ": ")
		if err != nil {
			log.Println(err)
			return
		}
	}

	dsn := libase.DsnInfo{
		Host:     *fHost,
		Port:     *fPort,
		Username: *fUser,
		Password: *fPass,
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return
	}
	defer db.Close()

	// test the database connection
	err = db.Ping()
	if err != nil {
		log.Printf("Pinging the server failed: %v", err)
		return
	}

	if len(flag.Args()) == 0 {
		return
	}

	subcmds := strings.Split(strings.Join(flag.Args(), " "), "--")
	for _, s := range subcmds {
		subcmd(db, strings.TrimSpace(s))
	}
}
