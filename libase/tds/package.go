package tds

import (
	"fmt"
	"io"
)

type Package interface {
	// ReadFrom reads bytes from the passed channel until either the
	// channel is closed or the package has all required information.
	// The read bytes are parsed into the package struct.
	ReadFrom(BytesChannel) error

	// WriteTo writes bytes to the passed channel until either the
	// channel is closed or the package has written all required
	// information.
	WriteTo(BytesChannel) error

	fmt.Stringer
}

func LookupPackage(token TDSToken) (Package, error) {
	switch token {
	case TDS_EED:
		return &EEDPackage{}, nil
	case TDS_ERROR:
		return &ErrorPackage{}, nil
	case TDS_LOGINACK:
		return &LoginAckPackage{}, nil
	case TDS_DONE:
		return &DonePackage{}, nil
	case TDS_DONEPROC:
		return &DoneProcPackage{}, nil
	case TDS_DONEINPROC:
		return &DoneInProcPackage{}, nil
	case TDS_MSG:
		return &MsgPackage{}, nil
	case TDS_PARAMFMT:
		return &ParamFmtPackage{}, nil
	case TDS_PARAMFMT2:
		return &ParamFmtPackage{wide: true}, nil
	case TDS_ROWFMT:
		return &RowFmtPackage{}, nil
	case TDS_ROWFMT2:
		return &RowFmtPackage{wide: true}, nil
	case TDS_PARAMS:
		return &ParamsPackage{}, nil
	case TDS_ROW:
		return &RowPackage{}, nil
	case TDS_CAPABILITY:
		return NewCapabilityPackage(nil, nil, nil)
	case TDS_ENVCHANGE:
		return &EnvChangePackage{}, nil
	case TDS_LANGUAGE:
		return &LanguagePackage{}, nil
	case TDS_ORDERBY:
		return &OrderByPackage{}, nil
	case TDS_ORDERBY2:
		return &OrderBy2Package{}, nil
	case TDS_RETURNSTATUS:
		return &ReturnStatusPackage{}, nil
	default:
		return NewTokenlessPackage(), nil
	}
}

func IsError(pkg Package) bool {
	switch pkg.(type) {
	case *EEDPackage, *ErrorPackage:
		return true
	}

	return false
}

func IsDone(pkg Package) bool {
	switch pkg.(type) {
	case *DonePackage:
		return true
	}

	return false
}

// The BytesChannel interface is accepted by all packages' WriteTo and
// ReadFrom methods.
type BytesChannel interface {
	// Position marks the index of the packet and index of the byte the
	// channel currently is at.
	// The position is considered volatile and only valid until the next
	// call to DiscardUntilPosition.
	Position() (int, int)
	SetPosition(int, int)
	DiscardUntilCurrentPosition()

	io.Reader
	io.Writer

	Bytes(n int) ([]byte, error)
	WriteBytes([]byte) error

	Byte() (byte, error)
	WriteByte(byte) error

	Uint8() (uint8, error)
	WriteUint8(uint8) error

	Int8() (int8, error)
	WriteInt8(int8) error

	Uint16() (uint16, error)
	WriteUint16(uint16) error

	Int16() (int16, error)
	WriteInt16(int16) error

	Uint32() (uint32, error)
	WriteUint32(uint32) error

	Int32() (int32, error)
	WriteInt32(int32) error

	Uint64() (uint64, error)
	WriteUint64(uint64) error

	Int64() (int64, error)
	WriteInt64(int64) error

	String(int) (string, error)
	WriteString(string) error
}
