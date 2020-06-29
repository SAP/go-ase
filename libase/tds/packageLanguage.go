package tds

import "fmt"

//go:generate stringer -type=LanguageStatus
type LanguageStatus int

const (
	TDS_LANGUAGE_NOARGS   LanguageStatus = 0x0
	TDS_LANGUAGE_HASARGS                 = 0x1
	TDS_LANG_BATCH_PARAMS                = 0x04
)

type LanguagePackage struct {
	Status LanguageStatus
	Cmd    string
}

func (pkg *LanguagePackage) ReadFrom(ch *channel) error {
	totalLength, err := ch.Uint32()
	if err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}

	status, err := ch.Byte()
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}
	pkg.Status = LanguageStatus(status)

	pkg.Cmd, err = ch.String(int(totalLength) - 1)
	if err != nil {
		return fmt.Errorf("failed to read language command: %w", err)
	}

	return nil
}

func (pkg *LanguagePackage) WriteTo(ch *channel) error {
	length := 1 + len(pkg.Cmd)
	err := ch.WriteUint32(uint32(length))
	if err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	err = ch.WriteByte(byte(pkg.Status))
	if err != nil {
		return fmt.Errorf("failed to write status: %w", err)
	}

	err = ch.WriteString(pkg.Cmd)
	if err != nil {
		return fmt.Errorf("failed to write language command: %w", err)
	}

	return nil
}

func (pkg LanguagePackage) String() string {
	return fmt.Sprintf("Language(%s): %s", pkg.Status, pkg.Cmd)
}
