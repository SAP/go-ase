// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"fmt"
	"io"

	"github.com/SAP/go-dblib/tds"
)

func handleDonePackage(pkg *tds.DonePackage) (bool, error) {
	if pkg.Status == tds.TDS_DONE_COUNT {
		return true, io.EOF
	}

	if pkg.Status&tds.TDS_DONE_ERROR == tds.TDS_DONE_ERROR {
		return true, fmt.Errorf("query failed with errors")
	}

	if pkg.Status&tds.TDS_DONE_MORE == tds.TDS_DONE_MORE ||
		pkg.Status&tds.TDS_DONE_INXACT == tds.TDS_DONE_INXACT ||
		pkg.Status&tds.TDS_DONE_PROC == tds.TDS_DONE_PROC {
		return false, nil
	}

	if pkg.Status == tds.TDS_DONE_FINAL {
		return true, io.EOF
	}

	return false, fmt.Errorf("%T with unrecognized Status: %s", pkg, pkg)
}
