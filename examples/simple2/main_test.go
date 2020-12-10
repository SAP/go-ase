// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Fatal(err)
	}
	// Output:
	// switching to master
	// dropping database newDB if exists
	// creating database newDB
	// creating table newDB..simple2_tab
	// inserting values into table newDB..simple2_tab
	// preparing statement
	// a: 5
	// teardown: switching to master
	// teardown: dropping database newDB
}
