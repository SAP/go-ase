package types

import (
	"reflect"
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
