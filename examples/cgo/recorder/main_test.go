// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Fatalf("Failed to execute example: %v", err)
	}
	// Output:
	// Opening database
	// Creating MessageRecorder
	// Registering handler with server message broker
	// Enable traceflag 3604
	// Received messages:
	// DBCC execution completed. If DBCC printed error messages, contact a user with System Administrator (SA) role.
	// Listing enabled traceflags
	// Received messages:
	// Active traceflags: 3604
	//
	// DBCC execution completed. If DBCC printed error messages, contact a user with System Administrator (SA) role.
}
