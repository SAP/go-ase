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
}

func (pkg *EEDPackage) ReadFrom(ch *channel) error {
	var err error
	pkg.Length, err = ch.Uint16()
	if err != nil {
		return err
	}

	pkg.MsgNumber, err = ch.Uint32()
	if err != nil {
		return err
	}

	pkg.State, err = ch.Uint8()
	if err != nil {
		return err
	}

	pkg.Class, err = ch.Uint8()
	if err != nil {
		return err
	}

	var sqlStateLen uint8
	sqlStateLen, err = ch.Uint8()
	if err != nil {
		return err
	}

	pkg.SQLState, err = ch.Bytes(int(sqlStateLen))
	if err != nil {
		return err
	}

	var status uint8
	status, err = ch.Uint8()
	if err != nil {
		return err
	}
	pkg.Status = EEDStatus(status)

	pkg.TranState, err = ch.Uint16()
	if err != nil {
		return err
	}

	var msgLength uint16
	msgLength, err = ch.Uint16()
	if err != nil {
		return err
	}

	pkg.Msg, err = ch.String(int(msgLength))
	if err != nil {
		return err
	}

	var serverLength uint8
	serverLength, err = ch.Uint8()
	if err != nil {
		return err
	}

	pkg.ServerName, err = ch.String(int(serverLength))
	if err != nil {
		return err
	}

	var procLength uint8
	procLength, err = ch.Uint8()
	if err != nil {
		return err
	}

	pkg.ProcName, err = ch.String(int(procLength))
	if err != nil {
		return err
	}

	pkg.LineNr, err = ch.Uint16()
	return err
}

func (pkg EEDPackage) WriteTo(ch *channel) error {
	return fmt.Errorf("not implemented")
}

func (pkg EEDPackage) String() string {
	return fmt.Sprintf("%d: %s", pkg.MsgNumber, pkg.Msg)
}
