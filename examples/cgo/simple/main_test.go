// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import "log"

func ExampleDoMain() {
	err := DoMain()
	if err != nil {
		log.Printf("Failed to execute example: %v", err)
	}
	// Output:
	//
	// Opening database
	// Creating table 'simple'
	// Writing a=2147483647, b='a string' to table
	// Querying values from table
	// Displaying results of query
	// | a          | b                              |
	// | 2147483647 | a string                       |
}
