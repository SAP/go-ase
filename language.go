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

func (c Conn) language(ctx context.Context, query string) (driver.Rows, driver.Result, error) {
	langPkg := &tds.LanguagePackage{
		Status: tds.TDS_LANGUAGE_NOARGS,
		Cmd:    query,
	}

	err := c.Channel.SendPackage(ctx, langPkg)
	if err != nil {
		return nil, nil, fmt.Errorf("error sending language command: %w", err)
	}

	return c.genericResults(ctx)
}
