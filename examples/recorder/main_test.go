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
	// Opening database
	// Enable traceflag 3604
	// Received messages on driver recorder:
	// driver: MsgNumber 5704: Changed client character set setting to 'utf8'.
	// driver: MsgNumber 5701: Changed database context to 'saptools'.
	// driver: MsgNumber 5703: Changed language setting to 'us_english'.
	// driver: MsgNumber 5704: Changed client character set setting to 'utf8'.
	// driver: MsgNumber 5701: Changed database context to 'saptools'.
	// driver: MsgNumber 5703: Changed language setting to 'us_english'.
	// driver: MsgNumber 2528: DBCC execution completed. If DBCC printed error messages, contact a user with System Administrator (SA) role.
	// Received messages on connector recorder:
	// connector: MsgNumber 5704: Changed client character set setting to 'utf8'.
	// connector: MsgNumber 5701: Changed database context to 'saptools'.
	// connector: MsgNumber 5703: Changed language setting to 'us_english'.
	// connector: MsgNumber 2528: DBCC execution completed. If DBCC printed error messages, contact a user with System Administrator (SA) role.
	// Opening second database
	// Received messages on driver recorder:
	// driver: MsgNumber 5704: Changed client character set setting to 'utf8'.
	// driver: MsgNumber 5701: Changed database context to 'saptools'.
	// driver: MsgNumber 5703: Changed language setting to 'us_english'.
	// driver: MsgNumber 5704: Changed client character set setting to 'utf8'.
	// driver: MsgNumber 5701: Changed database context to 'saptools'.
	// driver: MsgNumber 5703: Changed language setting to 'us_english'.
	// Received messages on connector recorder:
	// Received messages on connector recorder 2:
	// connector: MsgNumber 5704: Changed client character set setting to 'utf8'.
	// connector: MsgNumber 5701: Changed database context to 'saptools'.
	// connector: MsgNumber 5703: Changed language setting to 'us_english'.
}
