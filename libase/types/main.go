package types

import (
	"fmt"
	"reflect"
	"time"
)

//go:generate go run ./gen.go

type ASEType int

func Type(name string) ASEType {
	return string2type[name]
}

func (t ASEType) String() string {
	return type2string[t]
}

func (t ASEType) GoReflectType() reflect.Type {
	return type2reflect[t]
}

func (t ASEType) GoType() interface{} {
	return type2interface[t]
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
