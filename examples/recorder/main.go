// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how tds.EEDHooks can be utilized to record and
// inspect a number of tds.EEDPackages received during a transaction.
//
// go-ase supports three ways to add hooks:
//   1. driver-level
//   2. connector-level
//   3. connection-level
//
// Hooks at the driver-level receive EEDPackages from all connections
// and are added by calling ase.AddEEDHooks.
//
// Connector-level hooks receive EEDPackages from all connections opened
// through the connector they're attached to and are added by passing
// them to ase.NewConnectorWithHooks.
//
// Connection-level hooks receive EEDPackages only from their own
// connection and are added by passing them to ase.NewConnWithHooks.
//
// This example will explore driver- and connector-level hooks.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/tds"
)

const (
	exampleName  = "recorder"
	databaseName = exampleName + "DB"
	tableName    = databaseName + ".." + exampleName + "Table"
)

// Recorder is used to store EEDPackages as they are received by the
// driver.
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

// Reset deletes all stored EEDPackages.
func (rec *Recorder) Reset() {
	rec.Lock()
	defer rec.Unlock()

	rec.eeds = []tds.EEDPackage{}
}

// AddMessage is the callback function passed to the driver.
func (rec *Recorder) AddMessage(eed tds.EEDPackage) {
	rec.Lock()
	defer rec.Unlock()

	rec.eeds = append(rec.eeds, eed)
}

// LogMessages prints all stored messages to stdout.
func (rec *Recorder) LogMessages() {
	rec.RLock()
	defer rec.RUnlock()

	for _, eed := range rec.eeds {
		fmt.Printf("%s: MsgNumber %d: %s\n", rec.logprefix, eed.MsgNumber, eed.Msg)
	}
}

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("%s failed: %v", exampleName, err)
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

	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	// This is the recorder for the connector.
	// Every connection opened from this connector will send messages to
	// the connectorRecorder.
	connectorRecorder := NewRecorder("connector")
	connector, err := ase.NewConnectorWithHooks(info, nil, []tds.EEDHook{connectorRecorder.AddMessage})
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
	connector2, err := ase.NewConnectorWithHooks(info, nil, []tds.EEDHook{connectorRecorder2.AddMessage})
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
