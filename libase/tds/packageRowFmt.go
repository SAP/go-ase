package tds

import "fmt"

var _ Package = (*RowFmtPackage)(nil)

type RowFmtPackage struct {
	Fmts []FieldFmt
	// Wide differentiates TDS_ROWFMT from TDS_ROWFMT2 and considers the
	// length and status fields to be 4 bytes.
	// Otherwise the layout is exactly the same.
	wide bool
}

func (pkg *RowFmtPackage) ReadFrom(ch BytesChannel) error {
	totalLength, err := ch.Uint32()
	if err != nil {
		return fmt.Errorf("error reading length: %w", err)
	}

	colCount, err := ch.Uint16()
	if err != nil {
		return fmt.Errorf("error reading column count: %w", err)
	}
	readBytes := 2
	pkg.Fmts = make([]FieldFmt, colCount)

	for i := 0; i < int(colCount); i++ {
		fieldFmt, n, err := pkg.ReadFromField(ch)
		if err != nil {
			return fmt.Errorf("error reading column: %w", err)
		}
		pkg.Fmts[i] = fieldFmt
		readBytes += n
	}

	if readBytes != int(totalLength) {
		return fmt.Errorf("expected to read %d bytes, read %d bytes instead",
			totalLength, readBytes)
	}

	return nil
}

func (pkg *RowFmtPackage) ReadFromField(ch BytesChannel) (FieldFmt, int, error) {
	n := 0

	var label, catalogue, schema, table string
	if pkg.wide {
		labelLen, err := ch.Uint8()
		if err != nil {
			return nil, 0, fmt.Errorf("error reading label length: %w", err)
		}
		n++

		label, err = ch.String(int(labelLen))
		if err != nil {
			return nil, n, fmt.Errorf("error reading label: %w", err)
		}
		n += int(labelLen)

		catLen, err := ch.Uint8()
		if err != nil {
			return nil, n, fmt.Errorf("error reading catalogue length: %w", err)
		}
		n++

		catalogue, err = ch.String(int(catLen))
		if err != nil {
			return nil, n, fmt.Errorf("error reading catalogue name: %w", err)
		}
		n += int(catLen)

		schemaLen, err := ch.Uint8()
		if err != nil {
			return nil, n, fmt.Errorf("error reading schema length: %w", err)
		}
		n++

		schema, err = ch.String(int(schemaLen))
		if err != nil {
			return nil, n, fmt.Errorf("error reading schema: %w", err)
		}
		n += int(schemaLen)

		tableLen, err := ch.Uint8()
		if err != nil {
			return nil, n, fmt.Errorf("error reading table length: %w", err)
		}
		n++

		table, err = ch.String(int(tableLen))
		if err != nil {
			return nil, n, fmt.Errorf("error reading table: %w", err)
		}
		n += int(tableLen)
	}

	nameLength, err := ch.Uint8()
	if err != nil {
		return nil, 0, fmt.Errorf("error reading name length: %w", err)
	}
	n++

	name, err := ch.String(int(nameLength))
	if err != nil {
		return nil, n, fmt.Errorf("error reading name: %w", err)
	}
	n += int(nameLength)

	var status uint
	if pkg.wide {
		var status32 uint32
		status32, err = ch.Uint32()
		status = uint(status32)
		n += 4
	} else {
		var status8 uint8
		status8, err = ch.Uint8()
		status = uint(status8)
		n++
	}
	if err != nil {
		return nil, n, fmt.Errorf("error reading rowfmt status: %w", err)
	}

	userType, err := ch.Int32()
	if err != nil {
		return nil, n, fmt.Errorf("error reading user type: %w", err)
	}
	n += 4

	token, err := ch.Int8()
	if err != nil {
		return nil, n, fmt.Errorf("error reading data type: %w", err)
	}
	n++

	fieldFmt, err := LookupFieldFmt(DataType(token))
	if err != nil {
		return nil, n, fmt.Errorf("error preparing field format for token %s: %w", DataType(token), err)
	}

	fieldFmt.SetName(name)
	fieldFmt.SetStatus(uint(status))
	fieldFmt.SetUserType(userType)

	if pkg.wide {
		fieldFmt.SetColumnLabel(label)
		fieldFmt.SetCatalogue(catalogue)
		fieldFmt.SetSchema(schema)
		fieldFmt.SetTable(table)
	}

	readBytes, err := fieldFmt.ReadFrom(ch)
	if err != nil {
		return nil, n, fmt.Errorf("error reading param field format: %w", err)
	}
	n += readBytes

	localeLen, err := ch.Uint8()
	if err != nil {
		return nil, n, fmt.Errorf("error reading locale length: %w", err)
	}
	n++

	localeInfo, err := ch.String(int(localeLen))
	if err != nil {
		return nil, n, fmt.Errorf("error reading locale info: %w", err)
	}
	fieldFmt.SetLocaleInfo(localeInfo)
	n += int(localeLen)

	return fieldFmt, n, nil
}

func (pkg *RowFmtPackage) WriteTo(ch BytesChannel) error {
	return fmt.Errorf("not implemented")
}

func (pkg RowFmtPackage) String() string {
	wide := "nowide"
	if pkg.wide {
		wide = "wide"
	}

	s := fmt.Sprintf("%T(%s, %d): |", pkg, wide, len(pkg.Fmts))
	for _, f := range pkg.Fmts {
		s += fmt.Sprintf(" %s |", f.DataType())
	}

	return s
}
