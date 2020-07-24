package tds

import "fmt"

// HeaderOnlyPackage is used to communicate header-only packets using
// the same cummincation channels as regular token-based packages in
// go-ase.
type HeaderOnlyPackage struct {
	Header PacketHeader
}

func (pkg HeaderOnlyPackage) ReadFrom(ch BytesChannel) error {
	return fmt.Errorf("HeaderOnlyPackages cannot be read from a ByteChannel")
}

func (pkg HeaderOnlyPackage) WriteTo(ch BytesChannel) error {
	return fmt.Errorf("HeaderOnlyPackages cannot be written to a ByteChannel")
}

func (pkg HeaderOnlyPackage) String() string {
	return fmt.Sprintf("Header: %s", pkg.Header)
}
