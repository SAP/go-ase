package tds

import (
	"bytes"
	"fmt"
)

type TokenlessPackage struct {
	Data *bytes.Buffer
}

func NewTokenlessPackage() *TokenlessPackage {
	return &TokenlessPackage{
		Data: &bytes.Buffer{},
	}
}

func (pkg *TokenlessPackage) ReadFrom(ch *channel) error {
	_, err := pkg.Data.ReadFrom(ch)
	return err
}

func (pkg TokenlessPackage) WriteTo(ch *channel) error {
	return ch.WriteBytes(pkg.Data.Bytes())
}

func (pkg TokenlessPackage) String() string {
	return fmt.Sprintf("Tokenless(possibleToken=%x) %#v", pkg.Data.Bytes()[0], pkg)
}
