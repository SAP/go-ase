package cgo

import (
	"database/sql/driver"
	"errors"
)

//keep track of rows affected after inserts and updates
type Result struct {
	rowsAffected int64
}

// Interface satisfaction checks
var _ driver.Result = Result{}

func (result Result) LastInsertId() (int64, error) {
	// TODO
	return -1, errors.New("Feature not supported")
}

func (result Result) RowsAffected() (int64, error) {
	return result.rowsAffected, nil
}
