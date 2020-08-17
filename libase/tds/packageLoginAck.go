// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import "fmt"

//go:generate stringer -type=LoginAckStatus
type LoginAckStatus uint8

const (
	TDS_LOG_SUCCEED LoginAckStatus = 5 + iota
	TDS_LOG_FAIL
	TDS_LOG_NEGOTIATE
)

type LoginAckPackage struct {
	Length         uint16
	Status         LoginAckStatus
	Version        *Version
	NameLength     uint8
	ProgramName    string
	ProgramVersion *Version
}

func (pkg *LoginAckPackage) ReadFrom(ch BytesChannel) error {
	var err error

	pkg.Length, err = ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}

	var status uint8
	status, err = ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Status = (LoginAckStatus)(status)

	var vers []byte
	vers, err = ch.Bytes(4)
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Version, err = NewVersion(vers)
	if err != nil {
		return err
	}

	pkg.NameLength, err = ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}

	pkg.ProgramName, err = ch.String(int(pkg.NameLength))
	if err != nil {
		return ErrNotEnoughBytes
	}

	vers, err = ch.Bytes(4)
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.ProgramVersion, err = NewVersion(vers)

	return err
}

func (pkg LoginAckPackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteByte(byte(TDS_LOGINACK))
	if err != nil {
		return fmt.Errorf("failed to write TDS Token %s: %w", TDS_LOGINACK, err)
	}

	err = ch.WriteUint16(pkg.Length)
	if err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	err = ch.WriteUint8(uint8(pkg.Status))
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteBytes(pkg.Version.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteUint8(pkg.NameLength)
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteString(pkg.ProgramName)
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteBytes(pkg.ProgramVersion.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	return nil
}

func (pkg LoginAckPackage) String() string {
	return fmt.Sprintf("%T(%s)", pkg, pkg.Status)
}
