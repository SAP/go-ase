package types

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

var (
	_ driver.Valuer = (*NullTime)(nil)
	_ sql.Scanner   = (*NullTime)(nil)
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nullTime *NullTime) Scan(value interface{}) error {
	if value == nil {
		nullTime.Time = time.Time{}
		nullTime.Valid = false
		return nil
	}

	nullTime.Time, nullTime.Valid = value.(time.Time)
	return nil
}

func (nullTime NullTime) Value() (driver.Value, error) {
	if !nullTime.Valid {
		return nil, nil
	}

	return nullTime.Time, nil
}
