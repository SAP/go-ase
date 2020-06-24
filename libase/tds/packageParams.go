package tds

import (
	"fmt"
	"io/ioutil"
)

type ParamsPackage struct {
	paramFmt *ParamFmtPackage
	Params   []FieldData
}

func NewParamsPackage(params ...FieldData) *ParamsPackage {
	return &ParamsPackage{
		Params: params,
	}
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
	var err error
	for i, paramFmt := range pkg.paramFmt.Params {
		pkg.Params[i], err = LookupFieldData(paramFmt)
		if err != nil {
			return fmt.Errorf("error copying field: %w", err)
		}
	}

	return nil
}

func (pkg *ParamsPackage) ReadFrom(ch *channel) error {

	for i, param := range pkg.Params {
		err := param.ReadFrom(ch)
		if err != nil {
			return fmt.Errorf("error occurred reading param field %d data: %w", i, err)
		}
	}

	return nil
}

func (pkg ParamsPackage) WriteTo(ch *channel) error {
	err := ch.WriteByte(byte(TDS_PARAMS))
	if err != nil {
		return fmt.Errorf("error ocurred writing TDS token %s: %w", TDS_PARAMS, err)
	}

	for i, param := range pkg.Params {
		if err := param.WriteTo(ch); err != nil {
			return fmt.Errorf("error occurred writing param field %d data: %w", i, err)
		}
	}
	return nil
}

func (pkg ParamsPackage) String() string {
	return fmt.Sprintf("PARAMS(%d): ", len(pkg.Params))
}

func (pkg ParamsPackage) MultiString() []string {
	ret := make([]string, (len(pkg.Params) * 2))
	n := 0
	for _, param := range pkg.Params {
		ret[n] = fmt.Sprintf("%#v", param)

		stdoutCh := newChannel()
		param.WriteTo(stdoutCh)
		stdoutCh.Close()
		bs, _ := ioutil.ReadAll(stdoutCh)
		ret[n+1] = fmt.Sprintf("  Bytes(%d): %#v", len(bs), bs)

		n += 2
	}
	return ret
}
