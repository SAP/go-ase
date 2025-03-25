// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-dblib/tds"
)

func (c Conn) language(ctx context.Context, query string) (driver.Rows, driver.Result, error) {
	langPkg := &tds.LanguagePackage{
		Status: tds.TDS_LANGUAGE_NOARGS,
		Cmd:    query,
	}

	if err := c.Channel.SendPackage(ctx, langPkg); err != nil {
		return nil, nil, fmt.Errorf("error sending language command: %w", err)
	}

	return c.genericResults(ctx)
}
