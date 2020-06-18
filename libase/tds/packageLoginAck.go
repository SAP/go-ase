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
}

func (pkg *LoginAckPackage) ReadFrom(ch *channel) error {
	var err error

	pkg.Length, err = ch.Uint16()
	if err != nil {
		return err
	}

	var status uint8
	status, err = ch.Uint8()
	if err != nil {
		return err
	}
	pkg.Status = (LoginAckStatus)(status)

	var vers []byte
	vers, err = ch.Bytes(4)
	if err != nil {
		return err
	}
	pkg.TDSVersion, err = NewTDSVersion(vers)
	if err != nil {
		return err
	}

	pkg.NameLength, err = ch.Uint8()
	if err != nil {
		return err
	}

	pkg.ProgramName, err = ch.String(int(pkg.NameLength))
	if err != nil {
		return err
	}

	vers, err = ch.Bytes(4)
	if err != nil {
		return err
	}
	pkg.ProgramVersion, err = NewTDSVersion(vers)

	return err
}

func (pkg LoginAckPackage) WriteTo(ch *channel) error {
	err := ch.WriteUint16(pkg.Length)
	if err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	err = ch.WriteUint8(uint8(pkg.Status))
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteBytes(pkg.TDSVersion.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteUint8(pkg.NameLength)
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteString(pkg.ProgramName)
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	err = ch.WriteBytes(pkg.ProgramVersion.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write : %w", err)
	}

	return nil
}

func (pkg LoginAckPackage) String() string {
	return fmt.Sprintf("%s", pkg.Status)
}
