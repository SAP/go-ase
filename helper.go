// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"errors"
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

// finalize consumes all remaining packages in a communication using
// handleDonePackage.
// If any other package is received an error is returned.
func finalize(ctx context.Context, channel *tds.Channel) error {
	_, err := channel.NextPackageUntil(ctx, true, func(pkg tds.Package) (bool, error) {
		switch typed := pkg.(type) {
		case *tds.DonePackage:
			ok, err := handleDonePackage(typed)
			if err != nil {
				return true, err
			}
			return ok, nil
		default:
			return true, fmt.Errorf("go-ase: unhandled package type %T: %v", typed, typed)
		}
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
