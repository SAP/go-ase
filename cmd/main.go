package main

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/bgentry/speakeasy"
	_ "github.wdf.sap.corp/bssdb/go-ase/driver"
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
			panic(err)
		}
	}

	db, err := sql.Open("ase", "ase://"+*fUser+":"+pass+"@"+*fHost+":"+*fPort)
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
