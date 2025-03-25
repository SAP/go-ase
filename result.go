// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"database/sql/driver"
	"errors"
)

// Interface satisfaction checks
var _ driver.Result = (*Result)(nil)

// Result implements the driver.Result interface.
type Result struct {
	rowsAffected int64
}

// LastInsertId implements the driver.Result interface.
func (result Result) LastInsertId() (int64, error) {
	return -1, errors.New("not supported")
}

// RowsAffected implements the driver.Result interface.
func (result Result) RowsAffected() (int64, error) {
	return result.rowsAffected, nil
}
