package tds

import "fmt"

type ParamsPackage struct {
	paramFmt *ParamFmtPackage
	Params   []FieldData

	channelWrapper
}

func (pkg *ParamsPackage) LastPkg(other Package) error {
	switch other.(type) {
	case *ParamFmtPackage:
		pkg.paramFmt = other.(*ParamFmtPackage)
	case *ParamsPackage:
		pkg.paramFmt = other.(*ParamsPackage).paramFmt
	default:
		return fmt.Errorf("TDS_PARAMS received without preceeding TDS_PARAMFMT")
	}

	pkg.Params = make([]FieldData, len(pkg.paramFmt.Params))

	// Make copies of the formats to store data in
	for i, paramFmt := range pkg.paramFmt.Params {
		pkg.Params[i] = paramFmt.Copy()
	}

	return nil
}

func (pkg *ParamsPackage) ReadFrom(ch *channel) {
	pkg.ch = ch
	defer pkg.Finish()

	for i, param := range pkg.Params {
		err := param.readData(pkg.ch)
		if err != nil {
			pkg.err = fmt.Errorf("error occured reading param field %d data: %w", i, err)
			return
		}
	}
}

// TODO
func (pkg ParamsPackage) Packets() chan Packet {
	return nil
}

func (pkg ParamsPackage) String() string {
	s := fmt.Sprintf("PARAMS(%d): |", len(pkg.Params))
	for _, param := range pkg.Params {
		s += fmt.Sprintf(" %s |", param.Data())
	}
	return s
}
