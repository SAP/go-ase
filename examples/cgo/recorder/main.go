package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SAP/go-ase/cgo"
	libdsn "github.com/SAP/go-ase/libase/dsn"
)

func main() {
	err := DoMain()
	if err != nil {
		log.Fatal(err)
	}
}

func DoMain() error {
	dsn := libdsn.NewDsnInfoFromEnv("")

	fmt.Println("Opening database")
	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("Failed to open connection to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Creating MessageRecorder")
	recorder := cgo.NewMessageRecorder()
	fmt.Println("Registering handler with server message broker")
	cgo.GlobalServerMessageBroker.RegisterHandler(recorder.HandleMessage)

	fmt.Println("Calling dbcc")
	_, err = db.Exec("dbcc checkalloc")
	if err != nil {
		return err
	}

	fmt.Println("Received messages:")
	for _, line := range recorder.Text() {
		fmt.Print(line)
	}

	return nil
}
