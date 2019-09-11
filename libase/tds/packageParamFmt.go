package tds

import "fmt"

var _ Package = (*ParamFmtPackage)(nil)

type ParamFmtPackage struct {
	Length      uint16
	ParamsCount uint16
	Params      []FieldData

	Names       []string
	Stati       []uint8
	UserTypes   []int32
	LocaleInfos []string

	channelWrapper
}

func (pkg *ParamFmtPackage) ReadFrom(ch *channel) {
	pkg.ch = ch
	defer pkg.Finish()

	pkg.Length, pkg.err = pkg.ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.ParamsCount, pkg.err = pkg.ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.Params = make([]FieldData, pkg.ParamsCount)
	pkg.Names = make([]string, pkg.ParamsCount)
	pkg.Stati = make([]uint8, pkg.ParamsCount)
	pkg.UserTypes = make([]int32, pkg.ParamsCount)
	pkg.LocaleInfos = make([]string, pkg.ParamsCount)

	for i := 0; i < int(pkg.ParamsCount); i++ {
		nameLength, err := ch.Uint8()
		if err != nil {
			pkg.err = fmt.Errorf("failed to retrieve name length for field %d: %w", err)
			return
		}

		if nameLength > 0 {
			pkg.Names[i], err = ch.String(int(nameLength))
			if err != nil {
				pkg.err = fmt.Errorf("failed to retrieve name for field %d: %w", err)
				return
			}
		}

		pkg.Stati[i], err = ch.Uint8()
		if err != nil {
			pkg.err = fmt.Errorf("failed to retrieve status for field %d: %w", err)
			return
		}

		pkg.UserTypes[i], err = ch.Int32()
		if err != nil {
			pkg.err = fmt.Errorf("failed to retrieve usertype for field %d: %w", err)
			return
		}

		token, err := ch.Byte()
		if err != nil {
			pkg.err = fmt.Errorf("failed to retrieve token for field %d: %w",
				i, err)
			return
		}

		dataType := (DataType)(token)
		field, err := LookupFieldData(dataType)
		if err != nil {
			pkg.err = fmt.Errorf("error preparing field data struct for token %s: %w",
				dataType, err)
			return
		}

		err = field.readFormat(pkg.ch)
		if err != nil {
			pkg.err = fmt.Errorf("error occurred reading param field %d format: %w",
				i, err)
			return
		}

		localeLen, err := ch.Uint8()
		if err != nil {
			pkg.err = fmt.Errorf("error occurred reading locale length for field %d: %w", err)
			return
		}

		if localeLen > 0 {
			pkg.LocaleInfos[i], err = ch.String(int(localeLen))
			if err != nil {
				pkg.err = fmt.Errorf("error occurred reading locale info for field %d: %w", err)
				return
			}
		}

		pkg.Params[i] = field
	}
}

// TODO
func (pkg ParamFmtPackage) WriteTo(ch *channel) error {
	return nil
}

// TODO reconsider returning all params
func (pkg ParamFmtPackage) String() string {
	s := fmt.Sprintf("PARAMFMT(%d): |", pkg.Length)
	for _, param := range pkg.Params {
		s += fmt.Sprintf(" %s |", param.Type())
	}
	return s
}
