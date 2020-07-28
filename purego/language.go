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

	// Block until the first package of the response is available, then
	// read until an error occurs or all information has been read.
	var pkg tds.Package
	for pkg, err = c.Channel.NextPackage(ctx, true); err == nil; pkg, err = c.Channel.NextPackage(ctx, false) {
		switch typed := pkg.(type) {
		case *tds.RowFmtPackage:
			return &Rows{Conn: &c, RowFmt: typed}, nil, nil
		case *tds.DonePackage:
			result := &Result{}
			if typed.Status|tds.TDS_DONE_COUNT == tds.TDS_DONE_COUNT {
				result.rowsAffected = int64(typed.Count)
			}
			return nil, result, nil
		default:
			return nil, nil, fmt.Errorf("unhandled package type %T for language: %v", pkg, pkg)
		}
	}

	if err != nil {
		return nil, nil, fmt.Errorf("error reading package: %w", err)
	}

	return nil, nil, errors.New("no response from server")
}
