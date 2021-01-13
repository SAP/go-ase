// SPDX-FileCopyrightText: 2020 SAP SE
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
	// creating table subtransaction_tab
	// opening transaction
	// inserting values into subtransaction_tab
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
	// committing transaction
}
