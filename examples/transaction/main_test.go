// SPDX-FileCopyrightText: 2021 SAP SE
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
	// opening transaction
	// inserting values
	// preparing statement
	// executing prepared statement
	// a: 2147483647
	// b: a string
	// committing transaction
}
