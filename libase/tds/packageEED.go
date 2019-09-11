package tds

import "fmt"

type EEDStatus uint8

const (
	TDS_NO_EED      EEDStatus = 0x00
	TDS_EED_FOLLOWS           = 0x1
	TDS_EED_INFO              = 0x2
)

type EEDPackage struct {
	Length     uint16
	MsgNumber  uint32
	State      uint8
	Class      uint8
	SQLState   []byte
	Status     EEDStatus
	TranState  uint16
	Msg        string
	ServerName string
	ProcName   string
	LineNr     uint16

	channelWrapper
}

func (pkg *EEDPackage) ReadFrom(ch *channel) {
	pkg.ch = ch
	defer pkg.Finish()

	pkg.Length, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.MsgNumber, pkg.err = ch.Uint32()
	if pkg.err != nil {
		return
	}

	pkg.State, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.Class, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	var sqlStateLen uint8
	sqlStateLen, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.SQLState, pkg.err = ch.Bytes(int(sqlStateLen))
	if pkg.err != nil {
		return
	}

	var status uint8
	status, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}
	pkg.Status = EEDStatus(status)

	pkg.TranState, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	var msgLength uint16
	msgLength, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.Msg, pkg.err = ch.String(int(msgLength))
	if pkg.err != nil {
		return
	}

	var serverLength uint8
	serverLength, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ServerName, pkg.err = ch.String(int(serverLength))
	if pkg.err != nil {
		return
	}

	var procLength uint8
	procLength, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ProcName, pkg.err = ch.String(int(procLength))
	if pkg.err != nil {
		return
	}

	pkg.LineNr, pkg.err = ch.Uint16()
}

func (pkg EEDPackage) WriteTo(ch *channel) error {
	return nil
}

func (pkg EEDPackage) String() string {
	return fmt.Sprintf("%d: %s", pkg.MsgNumber, pkg.Msg)
}
