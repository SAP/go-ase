// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package cgo

import (
	"github.com/SAP/go-ase/libase/types"
)

//go:generate go run ./gen_types.go
type ASEType byte

// String satisfies the stringer interface.
func (t ASEType) String() string {
	s, ok := type2string[t]
	if !ok {
		return ""
	}
	return s
}

// ToDataType returns the equivalent types.DataType for an ASEType.
func (t ASEType) ToDataType() types.DataType {
	switch t {
	case BIGDATETIME:
		return types.BIGDATETIMEN
	case BIGINT:
		return types.INT8
	case BIGTIME:
		return types.BIGTIMEN
	case BINARY:
		return types.BINARY
	case BIT:
		return types.BIT
	case BLOB:
		return types.BLOB
	case BOUNDARY:
		return types.BOUNDARY
	case CHAR:
		return types.CHAR
	case DATE:
		return types.DATE
	case DATETIME:
		return types.DATETIME
	case DATETIME4:
		return types.SHORTDATE
	case DECIMAL:
		return types.DECN
	case FLOAT:
		return types.FLT8
	case IMAGE:
		return types.IMAGE
	case IMAGELOCATOR:
		// TODO
		return 0
	case INT:
		return types.INT4
	case LONG:
		return types.INT8
	case LONGBINARY:
		return types.LONGBINARY
	case LONGCHAR:
		return types.LONGCHAR
	case MONEY:
		return types.MONEY
	case MONEY4:
		return types.SHORTMONEY
	case NUMERIC:
		return types.NUMN
	case REAL:
		return types.FLT4
	case SENSITIVITY:
		return types.SENSITIVITY
	case SMALLINT:
		return types.INT2
	case TEXT:
		return types.TEXT
	case TEXTLOCATOR:
		// TODO
		return 0
	case TIME:
		return types.TIME
	case TINYINT:
		return types.INT1
	case UBIGINT:
		return types.UINT8
	case UINT:
		return types.UINT4
	case UNICHAR, UNITEXT:
		return types.UNITEXT
	case UNITEXTLOCATOR:
		// TODO
		return 0
	case USER:
		// TODO
		return 0
	case USHORT, USMALLINT:
		return types.UINT2
	case VARBINARY:
		return types.VARBINARY
	case VARCHAR:
		return types.VARCHAR
	case VOID:
		return types.VOID
	case XML:
		return types.XML
	default:
		return 0
	}
}
