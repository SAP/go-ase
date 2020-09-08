// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Fatal(err)
	}
	// Output:
	// opening transaction
	// creating table simple
	// inserting values into simple
	// preparing statement
	// executing prepared statement
	// a: 2147483647
	// b: a string
	// committing transaction
}
