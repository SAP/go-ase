package tds

import "fmt"

var _ Package = (*OrderByPackage)(nil)

type OrderByPackage struct {
	// Reference the previous RowFmt
	rowFmt      *RowFmtPackage
	ColumnOrder []int
}

func (pkg *OrderByPackage) LastPkg(other Package) error {
	if rowFmt, ok := other.(*RowFmtPackage); ok {
		pkg.rowFmt = rowFmt
		return nil
	}
	return fmt.Errorf("received package other than RowFmtPackage: %T", other)
}

func (pkg *OrderByPackage) ReadFrom(ch BytesChannel) error {
	columnCount, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}

	pkg.ColumnOrder = make([]int, int(columnCount))

	for i := range pkg.ColumnOrder {
		colNum, err := ch.Uint8()
		if err != nil {
			return ErrNotEnoughBytes
		}
		pkg.ColumnOrder[i] = int(colNum)
	}

	return nil
}

func (pkg OrderByPackage) WriteTo(ch BytesChannel) error {
	return fmt.Errorf("not implemented")
}

func (pkg OrderByPackage) String() string {
	return fmt.Sprintf("%T(%d): %v", pkg, len(pkg.ColumnOrder), pkg.ColumnOrder)
}

var _ Package = (*OrderBy2Package)(nil)

// OrderBy2Package is a superset of OrderByPackage and supports more
// than 255 columns.
type OrderBy2Package struct {
	OrderByPackage
}

func (pkg *OrderBy2Package) ReadFrom(ch BytesChannel) error {
	totalBytes, err := ch.Uint32()
	if err != nil {
		return fmt.Errorf("error reading byte length: %w", err)
	}

	columnCount, err := ch.Uint16()
	if err != nil {
		return fmt.Errorf("error reading column count: %w", err)
	}
	n := 2

	pkg.ColumnOrder = make([]int, int(columnCount))

	for i := range pkg.ColumnOrder {
		colNum, err := ch.Uint16()
		if err != nil {
			return fmt.Errorf("error reading column order for %d: %w", i, err)
		}
		n += 2
		pkg.ColumnOrder[i] = int(colNum)
	}

	if n != int(totalBytes) {
		return fmt.Errorf("expected to read %d bytes, read %d bytes instead",
			totalBytes, n)
	}

	return nil
}

func (pkg OrderBy2Package) WriteTo(ch BytesChannel) error {
	return fmt.Errorf("not implemented")
}

func (pkg OrderBy2Package) String() string {
	return fmt.Sprintf("%T(%d): %v", pkg, len(pkg.ColumnOrder), pkg.ColumnOrder)
}
