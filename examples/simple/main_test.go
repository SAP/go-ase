// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Printf("Failed to execute example: %v", err)
	}
	// Output:
	//
	// Opening database
	// Creating table 'simple_tab'
	// Writing a=2147483647, b='a string' to table
	// Querying values from table
	// Displaying results of query
	// | a          | b                              |
	// | 2147483647 | a string                       |
}
