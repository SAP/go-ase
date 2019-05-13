package types

import (
	"database/sql"
	"database/sql/driver"
)

var (
	_ driver.Valuer = (*NullBytes)(nil)
	_ sql.Scanner   = (*NullBytes)(nil)
)

type NullBytes struct {
	Bytes []byte
	Valid bool
}

func (nullBytes *NullBytes) Scan(value interface{}) error {
	nullBytes.Bytes, nullBytes.Valid = value.([]byte)
	return nil
}

func (nullBytes NullBytes) Value() (driver.Value, error) {
	if !nullBytes.Valid {
		return nil, nil
	}

	return nullBytes.Bytes, nil
}
