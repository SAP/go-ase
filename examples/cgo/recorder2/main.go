// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how a custom recorder can be implemented to
// process messages from the TDS server.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/libdsn"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("recorderexample: %v", err)
	}
}

func DoMain() error {
	recorder := &Recorder{}
	cgo.GlobalServerMessageBroker.RegisterHandler(recorder.HandleMessage)
	cgo.GlobalClientMessageBroker.RegisterHandler(recorder.HandleMessage)

	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	rows, err := db.Query("sp_adduser nologin")
	if err != nil {
		fmt.Println("Messages from ASE server:")
		for _, msg := range recorder.Messages {
			fmt.Printf("    %d: %s\n", msg.MessageNumber(), msg.Content())
		}
		return err
	}

	var returnStatus int
	for rows.Next() {
		err := rows.Scan(&returnStatus)
		if err != nil {
			return fmt.Errorf("error scanning return status: %w", err)
		}

		if returnStatus != 0 {
			fmt.Println("Messages from ASE server:")
			for _, msg := range recorder.Messages {
				fmt.Printf("    %d: %s\n", msg.MessageNumber(), msg.Content())
			}
			return fmt.Errorf("sp_adduser failed with return status %d", returnStatus)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows reported error: %w", err)
	}

	return nil
}

type Recorder struct {
	sync.RWMutex
	Messages []cgo.Message
}

func (rec *Recorder) HandleMessage(msg cgo.Message) {
	rec.Lock()
	defer rec.Unlock()

	rec.Messages = append(rec.Messages, msg)
}

func (rec *Recorder) Reset() {
	rec.Lock()
	defer rec.Unlock()

	rec.Messages = []cgo.Message{}
}
