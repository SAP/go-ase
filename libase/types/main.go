// Package types defines valid ASE types and auxiliary methods.
package types

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

//go:generate go run ./gen_types.go
//go:generate go run ./gen_nulls.go -type Binary -gotype []byte
//go:generate go run ./gen_nulls.go -type Time -gotype time.Time -import time

// ASEType reflects the data types ASE supports.
type ASEType int

// Type retuns an ASEType based on the name.
func Type(name string) ASEType {
	t, ok := string2type[name]
	if !ok {
		return ILLEGAL
	}
	return t
}

// String implements the Stringer interface.
func (t ASEType) String() string {
	s, ok := type2string[t]
	if !ok {
		return ""
	}
	return s
}

// GoReflectType returns the reflect.Type of the Go type used for the
// ASEType.
func (t ASEType) GoReflectType() reflect.Type {
	reflectType, ok := type2reflect[t]
	if !ok {
		return nil
	}
	return reflectType
}

// GoType returns the Go type used for the ASEType.
func (t ASEType) GoType() interface{} {
	goType, ok := type2interface[t]
	if !ok {
		return nil
	}
	return goType
}

// FromGoType returns the most fitting ASEType for the Go type.
func FromGoType(value interface{}) (ASEType, error) {
	switch value.(type) {
	case int64:
		return BIGINT, nil
	case float64:
		return FLOAT, nil
	case bool:
		return BIT, nil
	case []byte:
		return BINARY, nil
	case string:
		return CHAR, nil
	case time.Time:
		return BIGDATETIME, nil
	default:
		return ILLEGAL, fmt.Errorf("Invalid type for ASE: %v", value)
	}
}

type ValueConverter struct{}

var DefaultValueConverter = ValueConverter{}

func (conv ValueConverter) ConvertValue(v interface{}) (driver.Value, error) {
	if driver.IsValue(v) {
		return v, nil
	}

	switch value := v.(type) {
	case int:
		return int64(value), nil
	case uint:
		return uint64(value), nil
	}

	sv := reflect.TypeOf(v)
	for _, kind := range type2reflect {
		if kind == sv {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Unsupported type %T, a %s", v, sv.Kind())
}
