package tds

import (
	"fmt"
)

var _ Package = (*ParamsPackage)(nil)
var _ Package = (*RowPackage)(nil)

type ParamsPackage struct {
	paramFmt   *ParamFmtPackage
	rowFmt     *RowFmtPackage
	DataFields []FieldData
}

type RowPackage struct {
	ParamsPackage
}

func NewParamsPackage(data ...FieldData) *ParamsPackage {
	return &ParamsPackage{
		DataFields: data,
	}
}

func (pkg *ParamsPackage) LastPkg(other Package) error {
	switch otherPkg := other.(type) {
	case *ParamFmtPackage:
		pkg.paramFmt = otherPkg
	case *RowFmtPackage:
		pkg.rowFmt = otherPkg
	case *ParamsPackage:
		pkg.paramFmt = otherPkg.paramFmt
	case *RowPackage:
		pkg.rowFmt = otherPkg.rowFmt
	case *OrderByPackage:
		pkg.rowFmt = otherPkg.rowFmt
	case *OrderBy2Package:
		pkg.rowFmt = otherPkg.rowFmt
	default:
		return fmt.Errorf("TDS_PARAMS or TDS_ROW received without preceding TDS_PARAMFMT/2 or TDS_ROWFMT")
	}

	if pkg.DataFields != nil {
		// pkg.Datafiels has already been filled - this package was
		// created by the client and is being added to the message.
		return nil
	}

	var fieldFmts []FieldFmt
	if pkg.paramFmt != nil {
		fieldFmts = pkg.paramFmt.Fmts
	} else if pkg.rowFmt != nil {
		fieldFmts = pkg.rowFmt.Fmts
	} else {
		return fmt.Errorf("both paramFmt and rowFmt are nil")
	}

	pkg.DataFields = make([]FieldData, len(fieldFmts))

	// Make copies of the formats to store data in
	var err error
	for i, field := range fieldFmts {
		pkg.DataFields[i], err = LookupFieldData(field)
		if err != nil {
			return fmt.Errorf("error copying field: %w", err)
		}
	}

	return nil
}

func (pkg *ParamsPackage) ReadFrom(ch BytesChannel) error {
	for i, field := range pkg.DataFields {
		// TODO can the written byte count be validated?
		_, err := field.ReadFrom(ch)
		if err != nil {
			return fmt.Errorf("error occurred reading param field %d data (%s): %w",
				i, field.Format().DataType(), err)
		}
	}

	return nil
}

func (pkg ParamsPackage) WriteTo(ch BytesChannel) error {
	var token Token
	if pkg.paramFmt != nil {
		token = TDS_PARAMS
	} else if pkg.rowFmt != nil {
		token = TDS_ROW
	} else {
		return fmt.Errorf("both paramFmt and rowFmt are nil")
	}

	err := ch.WriteByte(byte(token))
	if err != nil {
		return fmt.Errorf("error ocurred writing TDS token %s: %w", token, err)
	}

	for i, field := range pkg.DataFields {
		if _, err := field.WriteTo(ch); err != nil {
			return fmt.Errorf("error occurred writing param field %d data: %w", i, err)
		}
	}
	return nil
}

func (pkg ParamsPackage) String() string {
	return fmt.Sprintf("%T(%d): ", pkg, len(pkg.DataFields))
}
