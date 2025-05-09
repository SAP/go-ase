// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration
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
