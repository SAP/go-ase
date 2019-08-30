/*

TDSConn is used to connect to a TDS server such as ASE.

The communication between the client and server is abstracted using
`Message`, which contain multiple packages.

A Package is a single type of information or instruction and may span
multiple Packets.
Packets are the equivalent of a PDU (Protocol Data Unit) in the official
TDS documentation and contain only the header and the data part.

Communication:
	Client				Server
	Message ->
		Package 1
			Packet/PDU 1
			Packet/PDU 2
		Package 2
			Packet 1
		Package 3
			Packet 1
						<- Message
							Package 1
								Packet 1
							Package 2
								Packet 1
							Package 3
								Packet 1
								Packet 2
							Done
	Message ->
		Package 1
			Packet 1

Messages are always sent and received in full or must be aborted.

*/
package tds
