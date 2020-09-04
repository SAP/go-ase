// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Printf("Failed to execute example: %v", err)
	}
	// Output:
	// Messages from ASE server:
	//     17231: No login with the specified name exists.
}
