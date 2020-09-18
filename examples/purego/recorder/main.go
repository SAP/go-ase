// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
	ase "github.com/SAP/go-ase/purego"
)

// This example shows how tds.EEDHooks can be utilized to access
// messages sent by ASE.

type Recorder struct {
	logprefix string
	eeds      []tds.EEDPackage
	sync.RWMutex
}

func NewRecorder(logprefix string) *Recorder {
	return &Recorder{
		logprefix: logprefix,
		eeds:      []tds.EEDPackage{},
		RWMutex:   sync.RWMutex{},
	}
}

func (rec *Recorder) Reset() {
	rec.Lock()
	defer rec.Unlock()

	rec.eeds = []tds.EEDPackage{}
}

func (rec *Recorder) AddMessage(eed tds.EEDPackage) {
	rec.Lock()
	defer rec.Unlock()

	rec.eeds = append(rec.eeds, eed)
}

func (rec *Recorder) LogMessages() {
	rec.RLock()
	defer rec.RUnlock()

	for _, eed := range rec.eeds {
		fmt.Printf("%s: MsgNumber %d: %s\n", rec.logprefix, eed.MsgNumber, eed.Msg)
	}
}

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("recorder: %v", err)
	}
}

func DoMain() error {
	// This is the recorder that is registered in the driver itself.
	// Every connection opened afterwards will send messages to the
	// driverRecorder.
	driverRecorder := NewRecorder("driver")
	if err := ase.AddEEDHooks(driverRecorder.AddMessage); err != nil {
		return fmt.Errorf("error adding EEDHook to driver: %w", err)
	}

	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	// This is the recorder for the connector.
	// Every connection opened from this connector will send messages to
	// the connectorRecorder.
	connectorRecorder := NewRecorder("connector")
	connector, err := ase.NewConnectorWithHooks(dsn, nil, []tds.EEDHook{connectorRecorder.AddMessage})
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	fmt.Println("Opening database")
	db := sql.OpenDB(connector)
	defer db.Close()

	fmt.Println("Enable traceflag 3604")
	if _, err := db.Exec("dbcc traceon(3604)"); err != nil {
		return fmt.Errorf("failed to enable traceflag 3604: %w", err)
	}

	fmt.Println("Received messages on driver recorder:")
	driverRecorder.LogMessages()
	driverRecorder.Reset()

	fmt.Println("Received messages on connector recorder:")
	connectorRecorder.LogMessages()
	connectorRecorder.Reset()

	// This is a second connectorRecorder - it will only receive
	// messages for connections opened from the second connector.
	// The driverRecorder will still receive messages from both
	// connector1 and connector2 connections.
	connectorRecorder2 := NewRecorder("connector")
	connector2, err := ase.NewConnectorWithHooks(dsn, nil, []tds.EEDHook{connectorRecorder2.AddMessage})
	if err != nil {
		return fmt.Errorf("failed to create connector2: %w", err)
	}

	fmt.Println("Opening second database")
	db2 := sql.OpenDB(connector2)
	defer db2.Close()

	rows, err := db.Query("select 'connector1'")
	if err != nil {
		return fmt.Errorf("failed to run sql statement in db: %w", err)
	}
	defer rows.Close()

	rows2, err := db2.Query("select 'connector2'")
	if err != nil {
		return fmt.Errorf("failed to run sql statement in db2: %w", err)
	}
	defer rows2.Close()

	fmt.Println("Received messages on driver recorder:")
	driverRecorder.LogMessages()
	driverRecorder.Reset()

	// This will print nothing, as no new messages have been recorded
	// on the recorder since database/sql reuses the previous
	// connection.
	fmt.Println("Received messages on connector recorder:")
	connectorRecorder.LogMessages()
	connectorRecorder.Reset()

	fmt.Println("Received messages on connector recorder 2:")
	connectorRecorder2.LogMessages()
	connectorRecorder2.Reset()

	return nil
}
