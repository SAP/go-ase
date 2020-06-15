package tds

import (
	"fmt"
)

type Package interface {
	// ReadFrom reads bytes from the passed channel until either the
	// channel is closed or the package has all required information.
	// The read bytes are parsed into the package struct.
	ReadFrom(*channel) error

	// WriteTo writes bytes to the passed channel until either the
	// channel is closed or the package has written all required
	// information.
	WriteTo(*channel) error

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
	case TDS_MSG:
		return &MsgPackage{}, nil
	case TDS_PARAMFMT:
		return &ParamFmtPackage{}, nil
	case TDS_PARAMS:
		return &ParamsPackage{}, nil
	}

	return nil, fmt.Errorf("no package type found for token '%s'", token)
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
