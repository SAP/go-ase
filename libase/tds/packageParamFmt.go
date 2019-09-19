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
}

func (pkg *ParamFmtPackage) ReadFrom(ch *channel) error {
	var err error

	pkg.Length, err = ch.Uint16()
	if err != nil {
		return err
	}

	pkg.ParamsCount, err = ch.Uint16()
	if err != nil {
		return err
	}

	pkg.Params = make([]FieldData, pkg.ParamsCount)
	pkg.Names = make([]string, pkg.ParamsCount)
	pkg.Stati = make([]uint8, pkg.ParamsCount)
	pkg.UserTypes = make([]int32, pkg.ParamsCount)
	pkg.LocaleInfos = make([]string, pkg.ParamsCount)

	for i := 0; i < int(pkg.ParamsCount); i++ {
		nameLength, err := ch.Uint8()
		if err != nil {
			return fmt.Errorf("failed to retrieve name length for field %d: %w", i, err)
		}

		if nameLength > 0 {
			pkg.Names[i], err = ch.String(int(nameLength))
			if err != nil {
				return fmt.Errorf("failed to retrieve name for field %d: %w", i, err)
			}
		}

		pkg.Stati[i], err = ch.Uint8()
		if err != nil {
			return fmt.Errorf("failed to retrieve status for field %d: %w", i, err)
		}

		pkg.UserTypes[i], err = ch.Int32()
		if err != nil {
			return fmt.Errorf("failed to retrieve usertype for field %d: %w", i, err)
		}

		token, err := ch.Byte()
		if err != nil {
			return fmt.Errorf("failed to retrieve token for field %d: %w", i, err)
		}

		dataType := (DataType)(token)
		field, err := LookupFieldData(dataType)
		if err != nil {
			return fmt.Errorf("error preparing field data struct for token %s: %w", dataType, err)
		}

		err = field.readFormat(pkg.ch)
		if err != nil {
			return fmt.Errorf("error occurred reading param field %d format: %w", i, err)
		}

		localeLen, err := ch.Uint8()
		if err != nil {
			return fmt.Errorf("error occurred reading locale length for field %d: %w", i, err)
		}

		if localeLen > 0 {
			pkg.LocaleInfos[i], err = ch.String(int(localeLen))
			if err != nil {
				return fmt.Errorf("error occurred reading locale info for field %d: %w", i, err)
			}
		}

		pkg.Params[i] = field
	}

	return nil
}

// TODO
func (pkg ParamFmtPackage) WriteTo(ch *channel) error {
	return fmt.Errorf("not implemented")
}

// TODO reconsider returning all params
func (pkg ParamFmtPackage) String() string {
	s := fmt.Sprintf("PARAMFMT(%d): |", pkg.Length)
	for _, param := range pkg.Params {
		s += fmt.Sprintf(" %s |", param.Type())
	}
	return s
}
