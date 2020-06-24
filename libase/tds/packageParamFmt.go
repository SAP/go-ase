package tds

import (
	"fmt"
)

var _ Package = (*ParamFmtPackage)(nil)

type ParamFmtPackage struct {
	Params []FieldFmt
	// Wide differentiates TDS_PARAMFMT from TDS_PARAMFMT2 and considers
	// the length and status fields to be 4 bytes
	// Otherwise the layout is exactly the same.
	wide bool
}

func NewParamFmtPackage(params ...FieldFmt) *ParamFmtPackage {
	return &ParamFmtPackage{Params: params}
}

func (pkg *ParamFmtPackage) ReadFrom(ch *channel) error {
	var err error

	// Read length
	if pkg.wide {
		_, err = ch.Uint32()
	} else {
		_, err = ch.Uint16()
	}
	if err != nil {
		return fmt.Errorf("failed to retrieve length: %w", err)
	}

	paramsCount, err := ch.Uint16()
	if err != nil {
		return err
	}

	pkg.Params = make([]FieldFmt, paramsCount)

	for i := 0; i < int(paramsCount); i++ {
		nameLength, err := ch.Uint8()
		if err != nil {
			return fmt.Errorf("failed to retrieve name length for field %d: %w", i, err)
		}

		var name string
		if nameLength > 0 {
			name, err = ch.String(int(nameLength))
			if err != nil {
				return fmt.Errorf("failed to retrieve name for field %d: %w", i, err)
			}
		}

		var status DataFieldStatus
		if pkg.wide {
			status32, err := ch.Uint32()
			if err != nil {
				return fmt.Errorf("failed to retrieve status for field %d: %w", i, err)
			}
			status = DataFieldStatus(status32)
		} else {
			status8, err := ch.Uint8()
			if err != nil {
				return fmt.Errorf("failed to retrieve status for field %d: %w", i, err)
			}
			status = DataFieldStatus(status8)
		}

		userType, err := ch.Int32()
		if err != nil {
			return fmt.Errorf("failed to retrieve usertype for field %d: %w", i, err)
		}

		token, err := ch.Byte()
		if err != nil {
			return fmt.Errorf("failed to retrieve token for field %d: %w", i, err)
		}

		dataType := (DataType)(token)

		fieldFmt, err := LookupFieldFmt(dataType)
		if err != nil {
			return fmt.Errorf("error preparing field format struct for token %s: %w", dataType, err)
		}

		// Set stored information on FieldData
		if len(name) > 0 {
			fieldFmt.SetName(name)
		}
		fieldFmt.SetStatus(status)
		fieldFmt.SetUserType(userType)

		err = fieldFmt.ReadFrom(ch)
		if err != nil {
			return fmt.Errorf("error occurred reading param field %d format: %w", i, err)
		}

		localeLen, err := ch.Uint8()
		if err != nil {
			return fmt.Errorf("error occurred reading locale length for field %d: %w", i, err)
		}

		if localeLen > 0 {
			localeInfo, err := ch.String(int(localeLen))
			if err != nil {
				return fmt.Errorf("error occurred reading locale info for field %d: %w", i, err)
			}
			fieldFmt.SetLocaleInfo(localeInfo)
		}

		pkg.Params[i] = fieldFmt
	}

	return nil
}

func (pkg ParamFmtPackage) WriteTo(ch *channel) error {
	var err error
	if pkg.wide {
		err = ch.WriteByte(byte(TDS_PARAMFMT2))
	} else {
		err = ch.WriteByte(byte(TDS_PARAMFMT))
	}
	if err != nil {
		return fmt.Errorf("error occurred writing TDS Token %s: %w", TDS_PARAMFMT, err)
	}

	// 2 bytes params count, 2 or 4 bytes length, x bytes for params
	length := 2
	if pkg.wide {
		length += 4
	} else {
		length += 2
	}
	for _, param := range pkg.Params {
		// 1 byte name length
		// x bytes name
		// 1 or 4 bytes status based on pkg.wide
		// 4 bytes usertype
		// 1 byte token
		// x bytes paramfmt (param.FormatByteLength())
		// 1 byte localeinfo length
		// x bytes localeinfo
		length += 7 + len(param.Name()) + param.FormatByteLength() + len(param.LocaleInfo())
		// status
		if pkg.wide {
			length += 4
		} else {
			length += 1
		}
	}

	if pkg.wide {
		if err := ch.WriteUint32(uint32(length)); err != nil {
			return fmt.Errorf("error occurred writing package length: %w", err)
		}
	} else {
		if err := ch.WriteUint16(uint16(length)); err != nil {
			return fmt.Errorf("error occurred writing package length: %w", err)
		}
	}

	if err := ch.WriteUint16(uint16(len(pkg.Params))); err != nil {
		return fmt.Errorf("error occurred writing params count: %w", err)
	}

	for i, param := range pkg.Params {
		if err := ch.WriteUint8(uint8(len(param.Name()))); err != nil {
			return fmt.Errorf("failed to write Name length for field %d: %w", i, err)
		}

		if len(param.Name()) > 0 {
			if err := ch.WriteString(param.Name()); err != nil {
				return fmt.Errorf("failed to write Name for field %d: %w", i, err)
			}
		}

		if pkg.wide {
			if err := ch.WriteUint32(uint32(param.Status())); err != nil {
				return fmt.Errorf("failed to write Status for field %d: %w", i, err)
			}
		} else {
			if err := ch.WriteUint8(uint8(param.Status())); err != nil {
				return fmt.Errorf("failed to write Status for field %d: %w", i, err)
			}
		}

		if err := ch.WriteInt32(param.UserType()); err != nil {
			return fmt.Errorf("failed to write UserType for field %d: %w", i, err)
		}

		if err := ch.WriteByte(byte(param.DataType())); err != nil {
			return fmt.Errorf("failed to write Token for field %d: %w", i, err)
		}

		if err := param.WriteTo(ch); err != nil {
			return fmt.Errorf("error writing ParamFmt field %d: %w", i, err)
		}

		if err := ch.WriteUint8(uint8(len(param.LocaleInfo()))); err != nil {
			return fmt.Errorf("failed to write LocaleInfo length for field %d: %w", i, err)
		}

		if len(param.LocaleInfo()) > 0 {
			if err := ch.WriteString(param.LocaleInfo()); err != nil {
				return fmt.Errorf("failed to write LocaleInfo for field %d: %w", i, err)
			}
		}
	}

	return nil
}

// TODO reconsider returning all params
func (pkg ParamFmtPackage) String() string {
	name := "PARAMFMT"
	if pkg.wide {
		name = "PARAMFMT2"
	}
	s := fmt.Sprintf("%s(%d): |", name, len(pkg.Params))
	for _, param := range pkg.Params {
		s += fmt.Sprintf(" %s |", param.DataType())
	}
	return s
}

func (pkg ParamFmtPackage) MultiString() []string {
	ret := make([]string, len(pkg.Params))
	for i, param := range pkg.Params {
		ret[i] = fmt.Sprintf("%#v", param)
	}
	return ret
}
