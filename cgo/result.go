package cgo

import "errors"

//keep track of rows affected after inserts and updates
type result struct {
	rowsAffected int64
}

func (result result) LastInsertId() (int64, error) {
	// TODO
	return -1, errors.New("Feature not supported")
}

func (result result) RowsAffected() (int64, error) {
	return result.rowsAffected, nil
}
