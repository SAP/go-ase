package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"

	_ "github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/libdsn"
)

func main() {
	err := DoMain()
	if err != nil {
		log.Printf("Failed: %v", err)
		os.Exit(1)
	}
}

func DoMain() error {
	dsn := libdsn.NewInfoFromEnv("")

	fmt.Println("Opening database")
	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("Failed to open connection to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("if object_id('simple') is not null drop table simple")
	if err != nil {
		return fmt.Errorf("Failed to drop table 'simple': %v", err)
	}

	fmt.Println("Creating table 'simple'")
	_, err = db.Exec("create table simple (a int, b char(30))")
	if err != nil {
		return fmt.Errorf("Failed to create table: %v", err)
	}

	fmt.Printf("Writing a=%d, b='a string' to table\n", math.MaxInt32)
	_, err = db.Exec("insert into simple values (?, \"?\")", math.MaxInt32, "a string")
	if err != nil {
		return fmt.Errorf("Failed to insert values: %v", err)
	}

	fmt.Println("Querying values from table")
	rows, err := db.Query("select * from simple")
	if err != nil {
		return fmt.Errorf("Querying failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Displaying results of query")
	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("Failed to retrieve column names: %v", err)
	}

	fmt.Printf("| %-10s | %-30s |\n", colNames[0], colNames[1])
	format := "| %-10d | %-30s |\n"

	var a int
	var b string

	for rows.Next() {
		err = rows.Scan(&a, &b)
		if err != nil {
			return fmt.Errorf("Failed to scan row: %v", err)
		}

		fmt.Printf(format, a, b)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("Error reading rows: %v", err)
	}

	return nil
}
