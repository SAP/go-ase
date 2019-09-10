package tds

import "fmt"

var _ Package = (*ParamFmtPackage)(nil)

type ParamFmtPackage struct {
	Length      uint16
	ParamsCount uint16
	Params      []FieldData

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

	for i := 0; i < int(pkg.ParamsCount); i++ {
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

		pkg.Params[i] = field
	}
}

// TODO
func (pkg ParamFmtPackage) Packets() chan Packet {
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
