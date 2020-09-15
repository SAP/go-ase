// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package main

func ExampleDoMain() {
	// This example is designed to fail to print the messages, as such
	// the error is ignored.
	DoMain()
	// Output:
	// Messages from ASE server:
	//     5701: Changed database context to 'saptools'.
	//
	//     5701: Changed database context to 'saptools'.
	//
	//     17231: No login with the specified name exists.
}
