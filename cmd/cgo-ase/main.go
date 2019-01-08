package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/SAP/go-ase/cgo"
	"github.com/bgentry/speakeasy"
)

var (
	fHost = flag.String("H", "localhost", "database hostname")
	fPort = flag.String("P", "4901", "database sql port")
	fUser = flag.String("u", "sa", "database user name")
	fPass = flag.String("p", "", "database user password")
)

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
		Database: *fDatabase,
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
		log.Printf("Pining the database failed: %v", err)
		return
	}

	if len(flag.Args()) == 0 {
		return
	}

	subcmd := flag.Args()[0]
	remainder := flag.Args()[1:]

	switch subcmd {
	case "exec":
		result, err := db.Exec(strings.Join(remainder, " "))
		if err != nil {
			log.Printf("Executing the statement failed: %v", err)
			return
		}

		affectedRows, err := result.RowsAffected()
		if err != nil {
			log.Printf("Retrieving the affected rows failed: %v", err)
			return
		}
		fmt.Printf("Rows affected: %d\n", affectedRows)
	case "query":
		rows, err := db.Query(strings.Join(remainder, " "))
		if err != nil {
			log.Printf("Query failed: %v", err)
			return
		}
		defer rows.Close()

		colNames, err := rows.Columns()
		if err != nil {
			log.Printf("Failed to retrieve column names: %v", err)
			return
		}
		fmt.Printf("|")
		for _, colName := range colNames {
			fmt.Printf(" %s |", colName)
		}
		fmt.Printf("\n")

		var a int
		var b string

		for rows.Next() {
			err := rows.Scan(&a, &b)
			if err != nil {
				log.Printf("Error retrieving rows: %v", err)
				return
			}
			fmt.Printf("| %d | %s |\n", a, b)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error preparing rows: %v", err)
			return
		}
	}
}
