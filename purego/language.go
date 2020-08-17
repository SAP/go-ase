// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql/driver"
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

	// TODO handle msg?

	pkg, err := c.Channel.NextPackage(true)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading package: %w", err)
	}

	// TODO handle result
	switch typed := pkg.(type) {
	case *tds.RowFmtPackage:
		return &Rows{Conn: &c, RowFmt: typed}, nil, nil
	default:
		return nil, nil, fmt.Errorf("unhandled package type %T for language: %v", pkg, pkg)
	}
}
