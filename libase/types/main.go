package types

import (
	"fmt"
	"reflect"
	"time"
)

//go:generate go run ./gen.go

type ASEType int

func Type(name string) ASEType {
	t, ok := string2type[name]
	if !ok {
		return ILLEGAL
	}
	return t
}

func (t ASEType) String() string {
	s, ok := type2string[t]
	if !ok {
		return ""
	}
	return s
}

func (t ASEType) GoReflectType() reflect.Type {
	t, ok := type2reflect[t]
	if !ok {
		return nil
	}
	return t
}

func (t ASEType) GoType() interface{} {
	t, ok := type2interface[t]
	if !ok {
		return nil
	}
	return t
}

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
