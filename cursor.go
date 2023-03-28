// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/SAP/go-dblib/namepool"
	"github.com/SAP/go-dblib/tds"
)

var (
	cursorPool = namepool.Pool("cursor%d")
)

// Cursor is used to interact with a cursor on the ASE server.
type Cursor struct {
	conn *Conn

	poolName *namepool.Name
	hasArgs  bool

	cursorID int32
	name     string
	// currently unused
	// tableName string

	currentRow int32
	totalRows  int32

	stmt *Stmt

	paramFmt *tds.ParamFmtPackage
	rowFmt   *tds.RowFmtPackage

	// When a CursorRows is created from a Cursor CursorRows.Close is
	// executed before the Cursor.Close.
	// Since CursorRows.Close needs to call Cursor.Close to correctly
	// deallocate a cursor when used through database/sql a flag is
	// required for direct use to prevent duplicated deallocation
	// requests.
	closed bool
}

// NewCursor creates a new cursor.
//
// NewCursor is a wrapper around NewCursorWithValues that converts
// arguments into driver.NamedValues.
func (c *Conn) NewCursor(ctx context.Context, query string, args ...interface{}) (*Cursor, error) {
	valueArgs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		valueArgs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   arg,
		}
	}

	return c.NewCursorWithValues(ctx, query, valueArgs)
}

// NewCursorWithValues creates a new cursor.
func (c *Conn) NewCursorWithValues(ctx context.Context, query string, args []driver.NamedValue) (*Cursor, error) {
	cursor := new(Cursor)
	cursor.conn = c

	if err := cursor.allocateOnServer(ctx, query, args); err != nil {
		return nil, fmt.Errorf("go-ase: error allocating cursor on server: %w", err)
	}

	return cursor, nil
}

// allocateOnServer allocates the cursor on the TDS server.
func (cursor *Cursor) allocateOnServer(ctx context.Context, query string, args []driver.NamedValue) error {
	cursor.poolName = cursorPool.Acquire()
	cursor.hasArgs = len(args) > 0

	if cursor.hasArgs {
		// cursor has argument, prepare statement
		stmt, err := cursor.conn.NewStmt(ctx, cursor.poolName.String(), query, true)
		if err != nil {
			return fmt.Errorf("error creating stmt: %w", err)
		}
		// The cursor needs a reference to the statement to close it.
		// The statement needs a reference to the cursor to mark it as
		// already closed if the server decides to close the cursor
		// immediately instead of waiting on being instructed to do so.
		cursor.stmt = stmt
		stmt.cursor = cursor
	}

	// If a cursor has arguments a statement with the query must be
	// prepared using the query. The name of the new statement is used
	// as the 'query' to reference the statement.
	cursorQuery := query
	if cursor.hasArgs {
		cursorQuery = cursor.poolName.String()
	}

	// Declare cursor.
	declarePkg, err := tds.NewCurDeclarePackage(cursor.poolName.String(), cursorQuery,
		tds.TDS_CUR_DSTAT_UNUSED,
		tds.TDS_CUR_DOPT_UNUSED,
	)
	if err != nil {
		return fmt.Errorf("could not create CurDeclarePackage: %w", err)
	}
	if cursor.hasArgs {
		declarePkg.Options |= tds.TDS_CUR_DOPT_DYNAMIC
	}

	if err := cursor.conn.Channel.SendPackage(ctx, declarePkg); err != nil {
		return fmt.Errorf("error sending CurDeclarePackage: %w", err)
	}

	_, err = cursor.conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) (bool, error) {
		switch typed := pkg.(type) {
		case *tds.DynamicPackage:
			if typed.Type&tds.TDS_DYN_ACK != tds.TDS_DYN_ACK {
				return true, fmt.Errorf("received DynamicPackage without type TDS_DYN_ACK: %s", typed)
			}

			return false, nil
		case *tds.CurInfoPackage:
			if typed.Command != tds.TDS_CUR_CMD_INFORM {
				return true, fmt.Errorf("received %T with command %s instead of TDS_CUR_CMD_INFORM",
					typed, typed.Command)
			}

			cursor.cursorID = typed.CursorID
			cursor.name = typed.Name
			cursor.currentRow = typed.RowNum
			cursor.totalRows = typed.TotalRows
			return false, nil
		case *tds.RowFmtPackage:
			cursor.rowFmt = typed
			return false, nil
		case *tds.ParamFmtPackage:
			cursor.paramFmt = typed
			return false, nil
		case *tds.DonePackage:
			ok, err := handleDonePackage(typed)
			if err != nil {
				return true, err
			}
			return ok, nil
		default:
			return true, fmt.Errorf("unhandled package type %T: %s", typed, typed)
		}
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error handling response to cursor creation: %w", err)
	}

	// Set how many rows should be sent by the TDS server per fetch.
	setFetchCount := &tds.CurInfoPackage{
		CursorID:  0,
		Name:      cursor.poolName.String(),
		Command:   tds.TDS_CUR_CMD_SETCURROWS,
		Status:    tds.TDS_CUR_ISTAT_ROWCNT,
		RowNum:    -1,
		TotalRows: 0,
		RowCount:  int32(cursor.conn.Info.CursorCacheRows),
	}

	if err := cursor.conn.Channel.SendPackage(ctx, setFetchCount); err != nil {
		return fmt.Errorf("error queueing CurInfoPackage to set fetch row count: %w", err)
	}

	_, err = cursor.conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) (bool, error) {
		switch typed := pkg.(type) {
		case *tds.DynamicPackage:
			if typed.Type&tds.TDS_DYN_ACK != tds.TDS_DYN_ACK {
				return true, fmt.Errorf("received DynamicPackage without type TDS_DYN_ACK: %s", typed)
			}

			return false, nil
		case *tds.CurInfoPackage:
			if typed.Command != tds.TDS_CUR_CMD_INFORM {
				return true, fmt.Errorf("received %T with command %s instead of TDS_CUR_CMD_INFORM",
					typed, typed.Command)
			}

			cursor.cursorID = typed.CursorID
			cursor.name = typed.Name
			cursor.currentRow = typed.RowNum
			cursor.totalRows = typed.TotalRows
			return false, nil
		case *tds.RowFmtPackage:
			cursor.rowFmt = typed
			return false, nil
		case *tds.ParamFmtPackage:
			cursor.paramFmt = typed
			return false, nil
		case *tds.DonePackage:
			ok, err := handleDonePackage(typed)
			if err != nil {
				return true, err
			}
			return ok, nil
		default:
			return true, fmt.Errorf("unhandled package type %T: %s", typed, typed)
		}
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error handling response to cursor creation: %w", err)
	}

	// Open cursor to read results.
	openPkg := &tds.CurOpenPackage{
		CursorID: cursor.cursorID,
		Name:     cursor.name,
	}

	if cursor.hasArgs {
		openPkg.Status = tds.TDS_CUR_OSTAT_HASARGS
	}

	if err := cursor.conn.Channel.QueuePackage(ctx, openPkg); err != nil {
		return fmt.Errorf("error queueing and sending CurOpenPackage: %w", err)
	}

	if cursor.hasArgs {
		if err := cursor.stmt.sendArgs(ctx, args); err != nil {
			return fmt.Errorf("error queueing arguments: %w", err)
		}
	}

	if err := cursor.conn.Channel.SendRemainingPackets(ctx); err != nil {
		return fmt.Errorf("error sending packages: %w", err)
	}

	_, err = cursor.conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) (bool, error) {
		switch typed := pkg.(type) {
		case *tds.CurInfoPackage:
			if typed.Command != tds.TDS_CUR_CMD_INFORM {
				return true, fmt.Errorf("received %T with command %s instead of TDS_CUR_CMD_INFORM",
					typed, typed.Command)
			}

			// TDS sends the rowcount before opening the cursor
			if typed.Status&tds.TDS_CUR_ISTAT_ROWCNT == tds.TDS_CUR_ISTAT_ROWCNT {
				return false, nil
			}

			if typed.Status&tds.TDS_CUR_ISTAT_OPEN != tds.TDS_CUR_ISTAT_OPEN {
				return true, fmt.Errorf("received %T without status TDS_CUR_ISTAT_OPEN", typed)
			}

			return false, nil
		case *tds.RowFmtPackage:
			cursor.rowFmt = typed
			return false, nil
		case *tds.ControlPackage:
			// TODO
			return false, nil
		case *tds.OrderByPackage, *tds.OrderBy2Package:
			return false, nil
		case *tds.DonePackage:
			ok, err := handleDonePackage(typed)
			if err != nil {
				return true, fmt.Errorf("go-ase: %w", err)
			}
			return ok, nil
		default:
			return true, fmt.Errorf("unhandled package type %T: %v", typed, typed)
		}
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error handling response to cursor opening: %w", err)
	}

	return nil
}

// Close closes the cursor.
func (cursor *Cursor) Close(ctx context.Context) error {
	defer cursorPool.Release(cursor.poolName)

	// If the cursor was already closed - e.g. because its result set
	// was exhausted and the CursorRows closed the cursor already, it
	// doesn't need to be closed again.
	// The statement will also be closed automatically by the server
	// once the result set is exhausted.
	if cursor.closed {
		return nil
	}

	if cursor.hasArgs {
		// cursor has an associated prepared statement
		if err := cursor.stmt.Close(); err != nil {
			return fmt.Errorf("go-ase: error closing stmt: %w", err)
		}

		// Return early if the cursor itself was automatically closed
		// after deallocating the associated statement.
		if cursor.closed {
			return nil
		}
	}

	closePkg := &tds.CurClosePackage{
		CursorID: cursor.cursorID,
		Name:     cursor.name,
		Options:  tds.TDS_CUR_COPT_DEALLOC,
	}

	if err := cursor.conn.Channel.SendPackage(ctx, closePkg); err != nil {
		return fmt.Errorf("go-ase: error sending CurClosePackage: %w", err)
	}

	// The deallocation request has been sent, mark cursor as closed.
	cursor.closed = true

	// TDS sometimes first sends an empty TDS_DONE_COUNT and
	// a TDS_DONE_FINAL, before sending the closing/deallocating
	// TDS_CUR_CMD_INFORM - sometimes though it directly sends the
	// closing/deallocating TDS_CUR_CMD_INFORM.
	// To handle this rxCurDealloc is set to true in the first stream
	// to only read the second stream if the cursor deallocation was not
	// confirmed by TDS.
	rxCurDealloc, err := cursor.closeReadResponse(ctx)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("go-ase: cursor deletion finished with error: %w", err)
	}

	if rxCurDealloc {
		return nil
	}

	if _, err := cursor.closeReadResponse(ctx); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("go-ase: cursor deletion finished with error: %w", err)
	}

	return nil
}

func (cursor *Cursor) closeReadResponse(ctx context.Context) (bool, error) {
	rxCurDealloc := false

	_, err := cursor.conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) (bool, error) {
		switch typed := pkg.(type) {
		case *tds.CurInfoPackage:
			if typed.Command != tds.TDS_CUR_CMD_INFORM {
				return true, fmt.Errorf("go-ase: received %T with command %s instead of TDS_CUR_CMD_INFORM",
					typed, typed.Command)
			}

			if typed.Status&tds.TDS_CUR_ISTAT_CLOSED != tds.TDS_CUR_ISTAT_CLOSED &&
				typed.Status&tds.TDS_CUR_ISTAT_DEALLOC != tds.TDS_CUR_ISTAT_DEALLOC {
				return true, fmt.Errorf("go-ase: received %T without status TDS_CUR_ISTAT_CLOSED or TDS_CUR_ISTAT_DEALLOC",
					typed)
			}

			if typed.Status&tds.TDS_CUR_ISTAT_DEALLOC == tds.TDS_CUR_ISTAT_DEALLOC {
				rxCurDealloc = true
			}

			return false, nil
		case *tds.DonePackage:
			// TDS sends an empty TDS_DONE_COUNT before TDS_DONE_FINAL
			if typed.Status == tds.TDS_DONE_COUNT {
				return false, nil
			}

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
		return rxCurDealloc, err
	}

	return rxCurDealloc, nil
}

// CursorID returns the ID assigned to the cursor by ASE.
func (cursor Cursor) CursorID() int {
	return int(cursor.cursorID)
}

// Fetch returns CursorRows to iterate over the rows selected by
// a cursor.
func (cursor *Cursor) Fetch(ctx context.Context) (*CursorRows, error) {
	return cursor.NewCursorRows()
}
