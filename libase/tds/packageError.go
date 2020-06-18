package tds

import "fmt"

type ErrorPackage struct {
	Length       uint16
	ErrorNumber  int32
	State        uint8
	Class        uint8
	MsgLength    uint16
	ErrorMsg     string
	ServerLength uint8
	ServerName   string
	ProcLength   uint8
	ProcName     string
	LineNr       uint16
}

func (pkg *ErrorPackage) ReadFrom(ch *channel) error {
	var err error
	pkg.Length, err = ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}

	pkg.ErrorNumber, err = ch.Int32()
	if err != nil {
		return fmt.Errorf("failed to read error number: %w", err)
	}

	pkg.MsgLength, err = ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read message length: %w", err)
	}

	pkg.ErrorMsg, err = ch.String(int(pkg.MsgLength))
	if err != nil {
		return fmt.Errorf("failed to read error message: %w", err)
	}

	pkg.ServerLength, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read servername length: %w", err)
	}

	pkg.ServerName, err = ch.String(int(pkg.ServerLength))
	if err != nil {
		return fmt.Errorf("failed to read servername: %w", err)
	}

	pkg.ProcLength, err = ch.Uint8()
	if err != nil {
		return fmt.Errorf("failed to read procname length: %w", err)
	}

	pkg.ProcName, err = ch.String(int(pkg.ProcLength))
	if err != nil {
		return fmt.Errorf("failed to read procname: %w", err)
	}

	pkg.LineNr, err = ch.Uint16()
	if err != nil {
		return fmt.Errorf("failed to read line nr: %w", err)
	}

	return nil
}

func (pkg ErrorPackage) WriteTo(ch *channel) error {
	return nil
}

func (pkg ErrorPackage) String() string {
	return fmt.Sprintf("%d: %s", pkg.ErrorNumber, pkg.ErrorMsg)
}
