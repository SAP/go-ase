package tds

import (
	"bytes"
	"io"
)

type TokenlessPackage struct {
	Data *bytes.Buffer
}

func (pkg *TokenlessPackage) ReadFrom(ch *channel) error {
	_, err := pkg.Data.ReadFrom(ch)
	return err
}

func (pkg TokenlessPackage) WriteTo(ch *channel) error {
	_, err := io.Copy(ch, pkg.Data)
	return err
}

// TODO
func (pkg TokenlessPackage) String() string {
	return ""
}
