package types

import (
	"fmt"
	"reflect"
	"time"
)

//go:generate go run ./gen.go

// ASEType reflects the data types ASE supports.
type ASEType int

// Type retuns an ASEType based on the name.
func Type(name string) ASEType {
	return string2type[name]
}

// String implements the Stringer interface.
func (t ASEType) String() string {
	return type2string[t]
}

// GoReflectType returns the reflect.Type of the Go type used for the
// ASEType.
func (t ASEType) GoReflectType() reflect.Type {
	return type2reflect[t]
}

// GoType returns the Go type used for the ASEType.
func (t ASEType) GoType() interface{} {
	return type2interface[t]
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
