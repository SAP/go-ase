package main

import (
	"database/sql"
	"fmt"
	"log"

	ase "github.com/SAP/go-ase/cgo"
	libdsn "github.com/SAP/go-ase/libase/dsn"
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}

// TODO

func doMain() error {
	dsn := libdsn.NewDsnInfoFromEnv("")

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return err
	}
	defer db.Close()

	recorder := ase.NewServerMessageRecorder()
	ase.GlobalServerMessageBroker.RegisterHandler(recorder.Handle)

	_, err = db.Exec("dbcc checkalloc")
	if err != nil {
		return err
	}

	_, lines, err := recorder.Text()
	if err != nil {
		return err
	}

	for _, line := range lines {
		fmt.Print(line)
	}

	return nil
}
