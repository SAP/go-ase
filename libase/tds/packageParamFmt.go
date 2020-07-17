package tds

import (
	"fmt"
)

var _ Package = (*ParamFmtPackage)(nil)

type ParamFmtPackage struct {
	Fmts []FieldFmt
	// Wide differentiates TDS_PARAMFMT from TDS_PARAMFMT2 and considers
	// the length and status fields to be 4 bytes.
	// Otherwise the layout is exactly the same.
	wide bool
}

func NewParamFmtPackage(wide bool, fmts ...FieldFmt) *ParamFmtPackage {
	return &ParamFmtPackage{wide: wide, Fmts: fmts}
}

func (pkg *ParamFmtPackage) ReadFrom(ch BytesChannel) error {
	// Read length
	totalBytes := 0
	if pkg.wide {
		length, err := ch.Uint32()
		if err != nil {
			return fmt.Errorf("failed to read length: %w", err)
		}
		totalBytes = int(uint(length))
	} else {
		length, err := ch.Uint16()
		if err != nil {
			return fmt.Errorf("failed to read length: %w", err)
		}
		totalBytes = int(uint(length))
	}

	n := 0

	paramsCount, err := ch.Uint16()
	if err != nil {
		return err
	}
	n += 2

	pkg.Fmts = make([]FieldFmt, int(paramsCount))

	for i := 0; i < int(paramsCount); i++ {
		param, readBytes, err := pkg.ReadFromField(ch)
		if err != nil {
			return err
		}

		// 1 namelength
		// x name
		// 4 or 1 status (wide)
		// 4 usertype
		// 1 token
		// x FormatByteLength
		// 1 locale len
		// x locale
		formatByteLength := 1 + len(param.Name()) + 1 + 4 + 1 + param.FormatByteLength() + 1 + len(param.LocaleInfo())
		if pkg.wide {
			formatByteLength += 3
		}

		if readBytes > formatByteLength {
			return fmt.Errorf("expected to read %d bytes for field %d, read %d bytes instead",
				formatByteLength, i, readBytes)
		}

		n += readBytes
		pkg.Fmts[i] = param
	}

	if n > totalBytes {
		return fmt.Errorf("expected to read %d bytes, read %d bytes instead",
			totalBytes, n)
	}

	return nil
}

func (pkg *ParamFmtPackage) ReadFromField(ch BytesChannel) (FieldFmt, int, error) {
	nameLength, err := ch.Uint8()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve name length: %w", err)
	}
	n := 1

	var name string
	name, err = ch.String(int(nameLength))
	if err != nil {
		return nil, n, fmt.Errorf("failed to retrieve name: %w", err)
	}
	n += int(nameLength)

	var status uint
	if pkg.wide {
		status32, err := ch.Uint32()
		if err != nil {
			return nil, n, fmt.Errorf("failed to retrieve status: %w", err)
		}
		status = uint(status32)
		n += 4
	} else {
		status8, err := ch.Uint8()
		if err != nil {
			return nil, n, fmt.Errorf("failed to retrieve status: %w", err)
		}
		status = uint(status8)
		n++
	}

	userType, err := ch.Int32()
	if err != nil {
		return nil, n, fmt.Errorf("failed to retrieve usertype: %w", err)
	}
	n += 4

	token, err := ch.Byte()
	if err != nil {
		return nil, n, fmt.Errorf("failed to retrieve token: %w", err)
	}
	n++

	fieldFmt, err := LookupFieldFmt(DataType(token))
	if err != nil {
		return nil, n, fmt.Errorf("error preparing field format %s: %w", DataType(token), err)
	}

	// Set stored information on FieldData
	fieldFmt.SetName(name)
	fieldFmt.SetStatus(status)
	fieldFmt.SetUserType(userType)

	n2, err := fieldFmt.ReadFrom(ch)
	if err != nil {
		return nil, n + n2, fmt.Errorf("error occurred reading param format: %w", err)
	}
	n += n2

	localeLen, err := ch.Uint8()
	if err != nil {
		return nil, n, fmt.Errorf("error occurred reading locale length: %w", err)
	}
	n++

	localeInfo, err := ch.String(int(localeLen))
	if err != nil {
		return nil, n, fmt.Errorf("error occurred reading locale info: %w", err)
	}
	fieldFmt.SetLocaleInfo(localeInfo)
	n += int(localeLen)

	return fieldFmt, n, nil
}

func (pkg ParamFmtPackage) WriteTo(ch BytesChannel) error {
	var err error
	if pkg.wide {
		err = ch.WriteByte(byte(TDS_PARAMFMT2))
	} else {
		err = ch.WriteByte(byte(TDS_PARAMFMT))
	}
	if err != nil {
		return fmt.Errorf("error occurred writing TDS Token %s: %w", TDS_PARAMFMT, err)
	}

	// 2 bytes params count, x bytes for params
	length := 2
	for _, field := range pkg.Fmts {
		// 1 byte name length
		// x bytes name
		// 1 or 4 bytes status based on pkg.wide
		// 4 bytes usertype
		// 1 byte token
		// x bytes paramfmt (param.FormatByteLength())
		// 1 byte localeinfo length
		// x bytes localeinfo
		length += 7 + len(field.Name()) + field.FormatByteLength() + len(field.LocaleInfo())
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

	if err := ch.WriteUint16(uint16(len(pkg.Fmts))); err != nil {
		return fmt.Errorf("error occurred writing params count: %w", err)
	}
	n := 2

	for i, field := range pkg.Fmts {
		writtenBytes, err := pkg.WriteToField(ch, field)
		if err != nil {
			return fmt.Errorf("error writing param %d: %w", i, err)
		}
		n += writtenBytes
	}

	if n > length {
		return fmt.Errorf("expected to write %d bytes, wrote %d bytes instead",
			length, n)
	}

	return nil
}

func (pkg ParamFmtPackage) WriteToField(ch BytesChannel, field FieldFmt) (int, error) {
	if err := ch.WriteUint8(uint8(len(field.Name()))); err != nil {
		return 0, fmt.Errorf("failed to write Name length: %w", err)
	}
	n := 1

	if err := ch.WriteString(field.Name()); err != nil {
		return n, fmt.Errorf("failed to write name: %w", err)
	}
	n += len(field.Name())

	if pkg.wide {
		if err := ch.WriteUint32(uint32(field.Status())); err != nil {
			return n, fmt.Errorf("failed to write status: %w", err)
		}
		n += 4
	} else {
		if err := ch.WriteUint8(uint8(field.Status())); err != nil {
			return n, fmt.Errorf("failed to write status: %w", err)
		}
		n += 1
	}

	if err := ch.WriteInt32(field.UserType()); err != nil {
		return n, fmt.Errorf("failed to write usertype: %w", err)
	}
	n += 4

	if err := ch.WriteByte(byte(field.DataType())); err != nil {
		return n, fmt.Errorf("failed to write token: %w", err)
	}
	n++

	n2, err := field.WriteTo(ch)
	if err != nil {
		return n, fmt.Errorf("error writing param format field: %w", err)
	}
	n += n2

	if err := ch.WriteUint8(uint8(len(field.LocaleInfo()))); err != nil {
		return n, fmt.Errorf("failed to write locale info length: %w", err)
	}
	n++

	if err := ch.WriteString(field.LocaleInfo()); err != nil {
		return n, fmt.Errorf("failed to write locale info: %w", err)
	}
	n += len(field.LocaleInfo())

	return n, nil
}

func (pkg ParamFmtPackage) String() string {
	wide := "nowide"
	if pkg.wide {
		wide = "wide"
	}
	s := fmt.Sprintf("%T(%s, %d): |", pkg, wide, len(pkg.Fmts))
	for _, field := range pkg.Fmts {
		s += fmt.Sprintf(" %s |", field.DataType())
	}
	return s
}

func (pkg ParamFmtPackage) MultiString() []string {
	ret := make([]string, len(pkg.Fmts))
	for i, field := range pkg.Fmts {
		ret[i] = fmt.Sprintf("%#v", field)
	}
	return ret
}
