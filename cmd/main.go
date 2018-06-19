package main

import (
	"database/sql"
	"fmt"

	_ "github.wdf.sap.corp/bssdb/go-ase/driver"
)

func main() {
	db, err := sql.Open("ase", "ase://user:pass@host:port")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// test the database connection
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
}
