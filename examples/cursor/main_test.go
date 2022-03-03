// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
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
	// opening cursor1
	// fetching cursor1
	// iterating over cursor1
	// a: 1
	// b: one
	// a: 2
	// b: two
	// a: 3
	// b: three
	// opening cursor2
	// fetching cursor2
	// iterating over cursor2
	// a: 2
	// b: two
}
