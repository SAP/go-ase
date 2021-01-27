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
	// before delete:
	// | a |
	// | 0 |
	// | 1 |
	// | 2 |
	// | 3 |
	// | 4 |
	// row: 0 a: 0
	// row: 1 a: 1
	// row: 2 a: 2
	// row: 3 a: 3
	// dropping row
	// row: 4 a: 4
	// after delete:
	// | a |
	// | 0 |
	// | 1 |
	// | 2 |
	// | 4 |
}
