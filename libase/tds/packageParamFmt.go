package tds

import (
	"fmt"
	"io/ioutil"
)

var _ Package = (*ParamFmtPackage)(nil)

type ParamFmtPackage struct {
	Params []FieldFmt
}

func NewParamFmtPackage(params ...FieldFmt) *ParamFmtPackage {
	return &ParamFmtPackage{Params: params}
}

func (pkg *ParamFmtPackage) ReadFrom(ch *channel) error {
	var err error

	// Read length - TODO use for validation
	_, err = ch.Uint16()
	if err != nil {
		return err
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

		status, err := ch.Uint8()
		if err != nil {
			return fmt.Errorf("failed to retrieve status for field %d: %w", i, err)
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
		fieldFmt.SetStatus(DataFieldStatus(status))
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
	// 2 bytes length, 2 bytes params count, x bytes for params
	length := 2 + 2
	for _, param := range pkg.Params {
		// 1 byte name length
		// x bytes name
		// 1 byte status
		// 4 bytes usertype
		// 1 byte token
		// x bytes paramfmt (param.FormatByteLength())
		// 1 byte localeinfo length
		// x bytes localeinfo
		length += 8 + len(param.Name()) + param.FormatByteLength() + len(param.LocaleInfo())
	}

	if err := ch.WriteUint16(uint16(length)); err != nil {
		return fmt.Errorf("error occurred writing package length: %w", err)
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

		if err := ch.WriteUint8(uint8(param.Status())); err != nil {
			return fmt.Errorf("failed to write Status for field %d: %w", i, err)
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
	s := fmt.Sprintf("PARAMFMT(%d): |", len(pkg.Params))
	for _, param := range pkg.Params {
		s += fmt.Sprintf(" %s |", param.DataType())
	}
	return s
}

func (pkg ParamFmtPackage) MultiString() []string {
	ret := make([]string, 1+(len(pkg.Params)*2))
	ret[0] = pkg.String()
	n := 1
	for _, param := range pkg.Params {
		ret[n] = fmt.Sprintf("  %#v", param)

		stdoutCh := newChannel()
		param.WriteTo(stdoutCh)
		stdoutCh.Close()
		bs, _ := ioutil.ReadAll(stdoutCh)
		ret[n+1] = fmt.Sprintf("    Bytes(%d): %#v", len(bs), bs)

		n += 2
	}
	return ret
}
