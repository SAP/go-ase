package tds

import (
	"fmt"
)

type Package interface {
	// ReadFrom reads bytes from the passed channel until either the
	// channel is closed or the package has all required information.
	// The read bytes are parsed into the package struct.
	ReadFrom(*channel)
	// Error returns an error encountered while parsing byes.
	Error() error
	// Finished returns true if the package has cannot parse any more
	// information.
	Finished() bool

	// // read packets from server into package
	// // TODO Remove
	// ReadFrom(packetsReader) (int64, error)

	// generate packets to send to server
	Packets() chan Packet

	fmt.Stringer
}

// TODO func to return a copy
var tokenToPackage = map[TDSToken]Package{
	TDS_ERROR: &ErrorPackage{},
}
