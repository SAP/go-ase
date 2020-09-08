// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"fmt"
)

//go:generate stringer -type=OptionCmd
type OptionCmd uint

const (
	TDS_OPT_SET OptionCmd = iota + 1
	TDS_OPT_DEFAULT
	TDS_OPT_LIST
	TDS_OPT_INFO
)

//go:generate stringer -type=OptionCmdOption
type OptionCmdOption uint

const (
	TDS_OPT_UNUSED OptionCmdOption = iota
	TDS_OPT_DATEFIRST
	TDS_OPT_TEXTSIZE
	TDS_OPT_STAT_TIME
	TDS_OPT_STAT_IO
	TDS_OPT_ROWCOUNT
	TDS_OPT_NATLANG
	TDS_OPT_DATEFORMAT
	TDS_OPT_ISOLATION
	TDS_OPT_AUTHON
	TDS_OPT_CHARSET
	TDS_OPT_PLAN
	TDS_OPT_ERRLVL
	TDS_OPT_SHOWPLAN
	TDS_OPT_NOEXEC
	TDS_OPT_ARITHIGNOREON
	TDS_OPT_ARITHABORTON
	TDS_OPT_PARSEONLY
	TDS_OPT_ESTIMATE
	TDS_OPT_GETDATA
	TDS_OPT_NOCOUNT
	TDS_OPT_FORCEPLAN
	TDS_OPT_FORMATONLY
	TDS_OPT_CHAINXACTS
	TDS_OPT_CURCLOSEONXACT
	TDS_OPT_FIPSFLAG
	TDS_OPT_RESTREES
	TDS_OPT_IDENTITYON
	TDS_OPT_CURREAD
	TDS_OPT_CURWRITE
	TDS_OPT_IDENTITYOFF
	TDS_OPT_AUTHOFF
	TDS_OPT_ANSINULL
	TDS_OPT_QUOTED_IDENT
	TDS_OPT_ANSIPERM
	TDS_OPT_STR_RTRUNC
	TDS_OPT_SORTMERGE
	TDS_OPT_JTC
	TDS_OPT_CLIENTREALNAME
	TDS_OPT_CLIENTHOSTNAME
	TDS_OPT_CLIENTAPPLNAME
	TDS_OPT_IDENTITYUPD_ON
	TDS_OPT_IDENTITYUPD_OFF
	TDS_OPT_NODATA
	TDS_OPT_CIPHERTEXT
	TDS_OPT_SHOW_FI
	TDS_OPT_HIDE_VCC
	TDS_OPT_LOBLOCATOR
	TDS_REQ_LOBLOCATOR
	TDS_OPT_LOBLOCATORFETCHSIZE
	TDS_OPT_ISOLATION_MODE OptionCmdOption = iota + 52
)

var _ Package = (*OptionCmdPackage)(nil)

type OptionCmdPackage struct {
	Cmd    OptionCmd
	Option OptionCmdOption
	// The type of OptionArg depends on the Option
	OptionArg []byte
}

func (pkg *OptionCmdPackage) ReadFrom(ch BytesChannel) error {
	_, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}

	cmd, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Cmd = OptionCmd(cmd)

	option, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Option = OptionCmdOption(option)

	argLength, err := ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	arg, err := ch.Bytes(int(argLength))
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.OptionArg = arg

	return nil
}

func (pkg OptionCmdPackage) WriteTo(ch BytesChannel) error {
	if err := ch.WriteByte(byte(TDS_OPTIONCMD)); err != nil {
		return err
	}

	// 1 cmd
	// 1 option
	// 1 argLength
	// x arg
	if err := ch.WriteUint16(uint16(3 + len(pkg.OptionArg))); err != nil {
		return err
	}

	if err := ch.WriteUint8(uint8(pkg.Cmd)); err != nil {
		return err
	}

	if err := ch.WriteUint8(uint8(pkg.Option)); err != nil {
		return err
	}

	if err := ch.WriteUint8(uint8(len(pkg.OptionArg))); err != nil {
		return err
	}

	if err := ch.WriteBytes(pkg.OptionArg); err != nil {
		return err
	}

	return nil
}

func (pkg OptionCmdPackage) String() string {
	return fmt.Sprintf("%T(%s, %s, %v)", pkg, pkg.Cmd, pkg.Option, pkg.OptionArg)
}
