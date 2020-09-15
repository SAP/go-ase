// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Printf("Failed to execute example: %v", err)
	}
	// Output:
	// Messages from ASE server:
	//     17231: No login with the specified name exists.
	// Messages from ASE server:
	//     156: Incorrect syntax near the keyword 'values'.
	//
	// Messages from ASE server:
	//     1809: CREATE DATABASE must be preceded by a 'USE master' command.  Check with your DBO <or a user with System Administrator (SA) role> if you do not have permission to USE master.
}
