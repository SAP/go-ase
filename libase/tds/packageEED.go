// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import "fmt"

//go:generate stringer -type=EEDStatus
type EEDStatus uint8

const (
	TDS_NO_EED      EEDStatus = 0x00
	TDS_EED_FOLLOWS EEDStatus = 0x1
	TDS_EED_INFO    EEDStatus = 0x2
)

type EEDPackage struct {
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

func (pkg *EEDPackage) ReadFrom(ch BytesChannel) error {
	length, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}

	pkg.MsgNumber, err = ch.Uint32()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n := 4

	pkg.State, err = ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	pkg.Class, err = ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	sqlStateLen, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	pkg.SQLState, err = ch.Bytes(int(sqlStateLen))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(sqlStateLen)

	status, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Status = EEDStatus(status)
	n++

	pkg.TranState, err = ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += 2

	msgLength, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += 2

	pkg.Msg, err = ch.String(int(msgLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(msgLength)

	serverLength, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	pkg.ServerName, err = ch.String(int(serverLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(serverLength)

	procLength, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n++

	pkg.ProcName, err = ch.String(int(procLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += int(procLength)

	pkg.LineNr, err = ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	n += 2

	if n != int(length) {
		return fmt.Errorf("expected to read %d bytes, read %d bytes instead", length, n)
	}

	return nil
}

func (pkg EEDPackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteByte(byte(TDS_EED))
	if err != nil {
		return fmt.Errorf("failed to write TDS Token %s: %w", TDS_EED, err)
	}

	// 4 msgnumber
	// 1 state
	// 1 class
	// x sqlstate
	// 1 status
	// 2 transtate
	// x msg
	// x servername
	// x procname
	// 2 linenr
	length := 11 + len(pkg.SQLState) + len(pkg.Msg) + len(pkg.ServerName) + len(pkg.ProcName)

	err = ch.WriteUint16(uint16(length))
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
	return fmt.Sprintf("%T(%s - %d: %s)", pkg, pkg.Status, pkg.MsgNumber, pkg.Msg)
}
