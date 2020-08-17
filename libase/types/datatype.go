// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package types

// DataType represents data types used in TDS-based applications.
//go:generate stringer -type=DataType
type DataType byte

const (
	BIGDATETIMEN DataType = 0xBB
	BIGTIMEN     DataType = 0xBC
	BINARY       DataType = 0x2D
	BIT          DataType = 0x32
	BLOB         DataType = 0x24
	BOUNDARY     DataType = 0x68
	CHAR         DataType = 0x2F
	DATE         DataType = 0x31
	DATEN        DataType = 0x7B
	DATETIME     DataType = 0x3D
	DATETIMEN    DataType = 0x6f
	DECN         DataType = 0x6A
	FLT4         DataType = 0x3B
	FLT8         DataType = 0x3E
	FLTN         DataType = 0x6D
	IMAGE        DataType = 0x22
	INT1         DataType = 0x30
	INT2         DataType = 0x34
	INT4         DataType = 0x38
	INT8         DataType = 0xBF
	INTN         DataType = 0x26
	LONGBINARY   DataType = 0xE1
	LONGCHAR     DataType = 0xAF
	MONEY        DataType = 0x3C
	MONEYN       DataType = 0x6E
	NUMN         DataType = 0x6C
	SENSITIVITY  DataType = 0x67
	SHORTDATE    DataType = 0x3A
	SHORTMONEY   DataType = 0x7A
	TEXT         DataType = 0x23
	TIME         DataType = 0x33
	TIMEN        DataType = 0x93
	UINT2        DataType = 0x41
	UINT4        DataType = 0x42
	UINT8        DataType = 0x43
	UINTN        DataType = 0x44
	UNITEXT      DataType = 0xAE
	VARBINARY    DataType = 0x25
	VARCHAR      DataType = 0x27
	VOID         DataType = 0x1f
	XML          DataType = 0xA3

	// Missing in tdspublic.h
	INTERVAL DataType = 0x2e
	SINT1    DataType = 0xb0

	// TDS usertypes
	USER_TEXT    DataType = 0x19
	USER_IMAGE   DataType = 0x20
	USER_UNITEXT DataType = 0x36
)
