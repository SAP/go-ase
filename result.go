// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"database/sql/driver"
	"errors"
)

var _ driver.Result = (*Result)(nil)

type Result struct {
	rowsAffected int64
}

func (result Result) LastInsertId() (int64, error) {
	return -1, errors.New("not supported")
}

func (result Result) RowsAffected() (int64, error) {
	return result.rowsAffected, nil
}
