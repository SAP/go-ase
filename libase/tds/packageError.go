package tds

import "fmt"

type ErrorPackage struct {
	Length       uint16
	ErrorNumber  int32
	State        uint8
	Class        uint8
	MsgLength    uint16
	ErrorMsg     string
	ServerLength uint8
	ServerName   string
	ProcLength   uint8
	ProcName     string
	LineNr       uint16

	channelWrapper
}

func (pkg *ErrorPackage) ReadFrom(ch *channel) {
	pkg.ch = ch
	defer pkg.Finish()

	pkg.Length, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.ErrorNumber, pkg.err = ch.Int32()
	if pkg.err != nil {
		return
	}

	pkg.MsgLength, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.ErrorMsg, pkg.err = ch.String(int(pkg.MsgLength))
	if pkg.err != nil {
		return
	}

	pkg.ServerLength, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ServerName, pkg.err = ch.String(int(pkg.ServerLength))
	if pkg.err != nil {
		return
	}

	pkg.ProcLength, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ProcName, pkg.err = ch.String(int(pkg.ProcLength))
	if pkg.err != nil {
		return
	}

	pkg.LineNr, pkg.err = ch.Uint16()
}

// TODO
func (pkg ErrorPackage) Packets() chan Packet {
	return nil
}

func (pkg ErrorPackage) String() string {
	return fmt.Sprintf("%d: %s", pkg.ErrorNumber, pkg.ErrorMsg)
}
