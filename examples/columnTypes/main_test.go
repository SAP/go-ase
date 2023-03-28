// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
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
	// querying table
	// column-name: a
	//   type: INT8
	//   length = 8
	//   nullable = false
	//   scan type: int64
	// column-name: b
	//   type: NUMN
	//   length = 15
	//   precision = 32
	//   scale = 0
	//   nullable = false
	//   scan type: *asetypes.Decimal
	// column-name: c
	//   type: DECN
	//   length = 8
	//   precision = 16
	//   scale = 2
	//   nullable = true
	//   scan type: *asetypes.Decimal
	// column-name: d
	//   type: CHAR
	//   length = 8
	//   nullable = false
	//   scan type: string
	// column-name: e
	//   type: VARCHAR
	//   length = 32
	//   nullable = true
	//   scan type: string
}
