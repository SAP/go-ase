package tds

import (
	"bytes"
)

type TokenlessPackage struct {
	Data *bytes.Buffer
}

func (pkg *TokenlessPackage) ReadFrom(ch *channel) error {
	_, err := pkg.Data.ReadFrom(ch)
	return err
}

func (pkg TokenlessPackage) WriteTo(ch *channel) error {
	return ch.WriteBytes(pkg.Data.Bytes())
}

// TODO
func (pkg TokenlessPackage) String() string {
	return ""
}
