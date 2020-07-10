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
		return fmt.Errorf("failed to read TDS token %s: %w", TDS_EED, err)
	}

	pkg.MsgNumber, err = ch.Uint32()
	if err != nil {
		return fmt.Errorf("failed to read message number: %w", err)
	}

	pkg.State, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	pkg.Class, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read class: %w", err)
	}

	var sqlStateLen uint8
	sqlStateLen, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read SQL state length: %w", err)
	}

	pkg.SQLState, err = ch.Bytes(int(sqlStateLen))
	if err != nil {
		return fmt.Errorf("failed to read SQL state: %w", err)
	}

	var status uint8
	status, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}
	pkg.Status = EEDStatus(status)

	pkg.TranState, err = ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read tran state: %w", err)
	}

	var msgLength uint16
	msgLength, err = ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read message length: %w", err)
	}

	pkg.Msg, err = ch.String(int(msgLength))
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	var serverLength uint8
	serverLength, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read server name length: %w", err)
	}

	pkg.ServerName, err = ch.String(int(serverLength))
	if err != nil {
		return fmt.Errorf("failed to read server name: %w", err)
	}

	var procLength uint8
	procLength, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read proc name length: %w", err)
	}

	pkg.ProcName, err = ch.String(int(procLength))
	if err != nil {
		return fmt.Errorf("failed to read proc name: %w", err)
	}

	pkg.LineNr, err = ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read line nr: %w", err)
	}

	return nil
}

func (pkg EEDPackage) WriteTo(ch *channel) error {
	err := ch.WriteByte(byte(TDS_EED))
	if err != nil {
		return fmt.Errorf("failed to write TDS Token %s: %w", TDS_EED, err)
	}

	err = ch.WriteUint16(pkg.Length)
	if err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	err = ch.WriteUint32(pkg.MsgNumber)
	if err != nil {
		return fmt.Errorf("failed to write message number: %w", err)
	}

	err = ch.WriteUint8(pkg.State)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	err = ch.WriteUint8(pkg.Class)
	if err != nil {
		return fmt.Errorf("failed to write class: %w", err)
	}

	err = ch.WriteUint8(uint8(len(pkg.SQLState)))
	if err != nil {
		return fmt.Errorf("failed to write SQL state len: %w", err)
	}

	err = ch.WriteBytes(pkg.SQLState)
	if err != nil {
		return fmt.Errorf("failed to write SQL state: %w", err)
	}

	err = ch.WriteByte(byte(pkg.Status))
	if err != nil {
		return fmt.Errorf("failed to write status: %w", err)
	}

	err = ch.WriteUint16(pkg.TranState)
	if err != nil {
		return fmt.Errorf("failed to write tran state: %w", err)
	}

	err = ch.WriteUint16(uint16(len(pkg.Msg)))
	if err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	err = ch.WriteString(pkg.Msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = ch.WriteUint8(uint8(len(pkg.ServerName)))
	if err != nil {
		return fmt.Errorf("failed to write server name length: %w", err)
	}

	err = ch.WriteString(pkg.ServerName)
	if err != nil {
		return fmt.Errorf("failed to write server name: %w", err)
	}

	err = ch.WriteUint8(uint8(len(pkg.ProcName)))
	if err != nil {
		return fmt.Errorf("failed to write proc name length: %w", err)
	}

	err = ch.WriteString(pkg.ProcName)
	if err != nil {
		return fmt.Errorf("failed to write proc name: %w", err)
	}

	err = ch.WriteUint16(pkg.LineNr)
	if err != nil {
		return fmt.Errorf("failed to write line nr: %w", err)
	}

	return nil
}

func (pkg EEDPackage) String() string {
	return fmt.Sprintf("%T(%d: %s)", pkg, pkg.MsgNumber, pkg.Msg)
}
