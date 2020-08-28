// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"database/sql/driver"
	"fmt"
	"reflect"
)

func (t DataType) ConvertValue(v interface{}) (driver.Value, error) {
	sv := reflect.ValueOf(v)

	// Return value as-is if the type already matches.
	if sv.Type() == ReflectTypes[t] {
		return v, nil
	}

	switch t {
	case INT1:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return uint8(sv.Int()), nil
		}
	case INT2:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return int16(sv.Int()), nil
		}
	case INT4:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return int32(sv.Int()), nil
		}
	case INT8, INTN:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return sv.Int(), nil
		}
	case UINT2:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return uint16(sv.Uint()), nil
		}
	case UINT4:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return uint32(sv.Uint()), nil
		}
	case UINT8, UINTN:
		switch sv.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return sv.Uint(), nil
		}
	case FLT4:
		switch sv.Kind() {
		case reflect.Float64:
			return float32(sv.Float()), nil
		}
	case FLT8:
		switch sv.Kind() {
		case reflect.Float32:
			return sv.Float(), nil
		}
	}

	return nil, fmt.Errorf("cannot convert %v (type %T) for %s", v, v, t)
}
