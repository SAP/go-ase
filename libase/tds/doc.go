// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

/*

Conn is used to connect to a TDS server such as ASE.

The communication between the client and server is abstracted using
`Conn` and `Channel`s.

The `Conn` reads payloads from the server in a separate goroutine and
routes them to their respective `Channel`.

The `Channel`s provide two main methods to send and receive data:
`NextPackage` to receive packages and `QueuePackage` to send packages.

A Package is a single type of information or instruction.

Communication:
	Client				Server
	SendPackage ->
		Package 1
		Package 2
		Package 3
						<- NextPackage
							Package 1
							Package 2
							Package 3
							Done
	SendPackage ->
		Package 1

The client/server communication is half-duplex - a string of consecutive
packages must be sent or received in full before a response or new
request can be transmitted.

*/
package tds
