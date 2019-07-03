package types

import (
	"database/sql"
	"database/sql/driver"
)

var (
	_ driver.Valuer = (*NullBinary)(nil)
	_ sql.Scanner   = (*NullBinary)(nil)
)

type NullBinary struct {
	Binary []byte
	Valid  bool
}

func (nullBinary *NullBinary) Scan(value interface{}) error {
	if value == nil {
		nullBinary.Binary = []byte{}
		nullBinary.Valid = false
		return nil
	}

	nullBinary.Binary, nullBinary.Valid = value.([]byte)
	return nil
}

func (nullBinary NullBinary) Value() (driver.Value, error) {
	if !nullBinary.Valid {
		return nil, nil
	}

	return nullBinary.Binary, nil
}
