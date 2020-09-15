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
	// opening transaction
	// creating table simple
	// inserting values into simple
	// reading table contents
	// a: 2147483647
	// b: a string
	// opening subtransaction
	// reading table contents with subtransaction changes
	// a: 2147483647
	// b: a string
	// a: 1000
	// b: another string
	// rolling back subtransaction
	// reading table contents after subtransaction rollback
	// a: 2147483647
	// b: a string
	// committing transaction
}
