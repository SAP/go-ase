package tds

import "fmt"

type ParamsPackage struct {
	paramFmt *ParamFmtPackage
	Params   []FieldData
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

func (pkg *ParamsPackage) ReadFrom(ch *channel) error {

	for i, param := range pkg.Params {
		err := param.readData(pkg.ch)
		if err != nil {
			return fmt.Errorf("error occurred reading param field %d data: %w", i, err)
		}
	}

	return nil
}

// TODO
func (pkg ParamsPackage) WriteTo(ch *channel) error {
	return fmt.Errorf("not implemented")
}

func (pkg ParamsPackage) String() string {
	s := fmt.Sprintf("PARAMS(%d): |", len(pkg.Params))
	for _, param := range pkg.Params {
		s += fmt.Sprintf(" %s |", param.Data())
	}
	return s
}
