package tds

import "fmt"

//go:generate stringer -type=LoginAckStatus
type LoginAckStatus uint8

const (
	TDS_LOG_SUCCEED LoginAckStatus = 5 + iota
	TDS_LOG_FAIL
	TDS_LOG_NEGOTIATE
)

type LoginAckPackage struct {
	Length         uint16
	Status         LoginAckStatus
	TDSVersion     *TDSVersion
	NameLength     uint8
	ProgramName    string
	ProgramVersion *TDSVersion

	channelWrapper
}

func (pkg *LoginAckPackage) ReadFrom(ch *channel) {
	pkg.ch = ch
	defer pkg.Finish()

	pkg.Length, pkg.err = pkg.ch.Uint16()
	if pkg.err != nil {
		return
	}

	var status uint8
	status, pkg.err = pkg.ch.Uint8()
	if pkg.err != nil {
		return
	}
	pkg.Status = (LoginAckStatus)(status)

	var vers []byte
	vers, pkg.err = pkg.ch.Bytes(4)
	if pkg.err != nil {
		return
	}
	pkg.TDSVersion, pkg.err = NewTDSVersion(vers)
	if pkg.err != nil {
		return
	}

	pkg.NameLength, pkg.err = pkg.ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ProgramName, pkg.err = pkg.ch.String(int(pkg.NameLength))
	if pkg.err != nil {
		return
	}

	vers, pkg.err = pkg.ch.Bytes(4)
	if pkg.err != nil {
		return
	}
	pkg.ProgramVersion, pkg.err = NewTDSVersion(vers)
}

func (pkg LoginAckPackage) WriteTo(ch *channel) error {
	return nil
}

func (pkg LoginAckPackage) String() string {
	return fmt.Sprintf("%s", pkg.Status)
}
