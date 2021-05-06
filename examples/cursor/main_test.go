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
	//
	// preparing table
	// inserted values:
	// | a          | b                              |
	// | 1          | one                            |
	// | 2          | two                            |
	// opening cursor1
	// fetching cursor1
	// iterating over cursor1
	// a: 1
	// b: one
	// a: 2
	// b: two
	// opening cursor2
	// fetching cursor2
	// iterating over cursor2
	// a: 2
	// b: two
}
