package tds

import (
	"bytes"
	"io"
)

type TokenlessPackage struct {
	Data     *bytes.Buffer
	err      error
	finished bool
}

func (pkg *TokenlessPackage) ReadFrom(ch *channel) {
	_, pkg.err = pkg.Data.ReadFrom(ch)
	pkg.finished = true
}

func (pkg TokenlessPackage) Error() error {
	return pkg.err
}

func (pkg TokenlessPackage) Finished() bool {
	return pkg.finished
}

func (pkg TokenlessPackage) WriteTo(ch *channel) error {
	_, err := io.Copy(ch, pkg.Data)
	return err
}

// TODO
func (pkg TokenlessPackage) String() string {
	return ""
}
