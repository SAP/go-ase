// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/SAP/go-ase/libase/tds"
)

func (c Conn) GenericExec(ctx context.Context, query string) (driver.Rows, driver.Result, error) {
	return c.language(ctx, query)
}

func (c Conn) language(ctx context.Context, query string) (driver.Rows, driver.Result, error) {
	langPkg := &tds.LanguagePackage{
		Status: tds.TDS_LANGUAGE_NOARGS,
		Cmd:    query,
	}

	err := c.Channel.SendPackage(ctx, langPkg)
	if err != nil {
		return nil, nil, fmt.Errorf("error sending language command: %w", err)
	}

	for {
		pkg, err := c.Channel.NextPackageUntil(ctx, true,
			func(pkg tds.Package) (bool, error) {
				switch typed := pkg.(type) {
				case *tds.RowFmtPackage:
					return true, nil
				case *tds.DonePackage:
					if typed.Status&tds.TDS_DONE_ERRROR == tds.TDS_DONE_ERRROR {
						return false, errors.New("received Done with error, transaction aborted")
					}
					return true, nil
				default:
					return false, nil
				}
			},
		)
		if err != nil {
			return nil, nil, err
		}

		switch typed := pkg.(type) {
		case *tds.RowFmtPackage:
			return &Rows{Conn: &c, RowFmt: typed}, nil, nil
		case *tds.DonePackage:
			result := &Result{}
			if typed.Status&tds.TDS_DONE_COUNT == tds.TDS_DONE_COUNT {
				result.rowsAffected = int64(typed.Count)
			}
			return nil, result, nil
		default:
			return nil, nil, fmt.Errorf("unhandled package type %T for language: %v", pkg, pkg)
		}
	}
}
