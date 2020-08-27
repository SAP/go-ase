// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"reflect"
	"time"
)

var ReflectTypes = map[DataType]reflect.Type{
	BIGDATETIMEN: reflect.TypeOf(time.Time{}),
	BIGTIMEN:     reflect.TypeOf(time.Time{}),
	BINARY:       reflect.SliceOf(reflect.TypeOf(byte(0))),
	BIT:          reflect.TypeOf(bool(false)),
	BLOB:         reflect.SliceOf(reflect.TypeOf(byte(0))),
	BOUNDARY:     nil,
	CHAR:         reflect.TypeOf(string("")),
	DATE:         reflect.TypeOf(time.Time{}),
	DATEN:        reflect.TypeOf(time.Time{}),
	DATETIME:     reflect.TypeOf(time.Time{}),
	DATETIMEN:    reflect.TypeOf(time.Time{}),
	DECN:         reflect.TypeOf(&Decimal{}),
	FLT4:         reflect.TypeOf(float32(0)),
	FLT8:         reflect.TypeOf(float64(0)),
	FLTN:         reflect.TypeOf(float64(0)),
	IMAGE:        reflect.SliceOf(reflect.TypeOf(byte(0))),
	INT1:         reflect.TypeOf(int8(0)),
	INT2:         reflect.TypeOf(int16(0)),
	INT4:         reflect.TypeOf(int32(0)),
	INT8:         reflect.TypeOf(int64(0)),
	INTN:         reflect.TypeOf(int64(0)),
	LONGBINARY:   reflect.SliceOf(reflect.TypeOf(byte(0))),
	LONGCHAR:     reflect.TypeOf(string("")),
	MONEY:        reflect.TypeOf(&Decimal{}),
	MONEYN:       reflect.TypeOf(&Decimal{}),
	NUMN:         reflect.TypeOf(&Decimal{}),
	SENSITIVITY:  nil,
	SHORTDATE:    reflect.TypeOf(time.Time{}),
	SHORTMONEY:   reflect.TypeOf(time.Time{}),
	TEXT:         reflect.TypeOf(string("")),
	TIME:         reflect.TypeOf(time.Time{}),
	TIMEN:        reflect.TypeOf(time.Time{}),
	UINT2:        reflect.TypeOf(uint16(0)),
	UINT4:        reflect.TypeOf(uint32(0)),
	UINT8:        reflect.TypeOf(uint64(0)),
	UINTN:        reflect.TypeOf(uint64(0)),
	UNITEXT:      reflect.TypeOf(string("")),
	VARBINARY:    reflect.SliceOf(reflect.TypeOf(byte(0))),
	VARCHAR:      reflect.TypeOf(string("")),
	VOID:         nil,
	XML:          reflect.SliceOf(reflect.TypeOf(byte(0))),

	INTERVAL: nil,
	SINT1:    nil,

	USER_TEXT:    nil,
	USER_IMAGE:   nil,
	USER_UNITEXT: nil,
}

func (t DataType) GoReflectType() reflect.Type {
	reflectType, ok := ReflectTypes[t]
	if !ok {
		return nil
	}
	return reflectType
}
