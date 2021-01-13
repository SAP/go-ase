// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
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
	// sp_adduser
	// Messages from ASE server:
	//     17231: No login with the specified name exists.
	// create table eed_example_tab
	// Messages from ASE server:
	//     156: Incorrect syntax near the keyword 'values'.
	// create database
	// Messages from ASE server:
	//     102: Incorrect syntax near 'database'.
}
