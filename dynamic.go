// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"

	"github.com/SAP/go-dblib"
	"github.com/SAP/go-dblib/namepool"
	"github.com/SAP/go-dblib/tds"
)

// Interface satisfaction checks.
var (
	_ driver.Stmt              = (*Stmt)(nil)
	_ driver.StmtExecContext   = (*Stmt)(nil)
	_ driver.StmtQueryContext  = (*Stmt)(nil)
	_ driver.NamedValueChecker = (*Stmt)(nil)

	stmtIdPool = namepool.Pool("stmt%d")
)

// Stmt implements the driver.Stmt interface.
type Stmt struct {
	conn *Conn

	stmtId *namepool.Name
	pkg    *tds.DynamicPackage

	paramFmt *tds.ParamFmtPackage
	rowFmt   *tds.RowFmtPackage
}

// Prepare implements the driver.Conn interface.
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

// PrepareContext implements the driver.ConnPrepareContext interface.
func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	// TODO option for create_proc
	return c.NewStmt(ctx, "", query, true)
}

// NewStmt creates a new statement.
func (c *Conn) NewStmt(ctx context.Context, name, query string, create_proc bool) (*Stmt, error) {
	stmt := &Stmt{conn: c}

	if name == "" {
		// TODO different pools for procs and prepares
		stmt.stmtId = stmtIdPool.Acquire()
		name = stmt.stmtId.Name()
	}

	stmt.pkg = tds.NewDynamicPackage(true)
	stmt.pkg.ID = name

	if create_proc {
		stmt.pkg.Stmt = fmt.Sprintf("create proc %s as %s", name, query)
	} else {
		stmt.pkg.Stmt = query
	}

	// Reset statement to default before proceeding
	stmt.Reset()

	if err := stmt.allocateOnServer(ctx); err != nil {
		return nil, fmt.Errorf("go-ase: error allocating dynamic statement '%s': %w", query, err)
	}

	return stmt, nil
}

// allocateOnServer communicates the allocation of the dynamic statement
// on the server and retrieves the input and output formats.
func (stmt *Stmt) allocateOnServer(ctx context.Context) error {
	stmt.pkg.Type = tds.TDS_DYN_PREPARE
	if err := stmt.conn.Channel.SendPackage(ctx, stmt.pkg); err != nil {
		return fmt.Errorf("error queueing dynamic prepare package: %w", err)
	}
	stmt.Reset()

	if err := stmt.recvDynAck(ctx); err != nil {
		return err
	}

	_, err := stmt.conn.Channel.NextPackageUntil(ctx, true,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.ParamFmtPackage:
				stmt.paramFmt = typed
				return false, nil
			case *tds.RowFmtPackage:
				stmt.rowFmt = typed
				return false, nil
			case *tds.DonePackage:
				ok, err := handleDonePackage(typed)
				if err != nil {
					return true, err
				}
				return ok, nil
			default:
				return false, fmt.Errorf("unexpected package received: %#v", typed)
			}
		},
	)
	if err != nil && !errors.Is(err, io.EOF) {
		stmt.close(ctx)
		return err
	}

	return nil
}

func (stmt Stmt) fieldFmts() ([]tds.FieldFmt, error) {
	if stmt.paramFmt != nil {
		return stmt.paramFmt.Fmts, nil
	} else if stmt.rowFmt != nil {
		return stmt.rowFmt.Fmts, nil
	}
	return nil, fmt.Errorf("bot paramFmt and rowFmt are unset")
}

// Reset resets a statement.
func (stmt *Stmt) Reset() {
	stmt.pkg.Type = tds.TDS_DYN_INVALID
	stmt.pkg.Status = tds.TDS_DYNAMIC_UNUSED
}

// Close implements the driver.Stmt interface.
func (stmt *Stmt) Close() error {
	return stmt.close(context.Background())
}

func (stmt *Stmt) close(ctx context.Context) error {
	if stmt.stmtId != nil {
		defer stmtIdPool.Release(stmt.stmtId)
	}

	// communicate deallocation with server
	// TODO option to not deallocate procs
	stmt.pkg.Type = tds.TDS_DYN_DEALLOC
	if err := stmt.conn.Channel.SendPackage(ctx, stmt.pkg); err != nil {
		return fmt.Errorf("error sending dealloc package: %w", err)
	}
	stmt.Reset()

	if err := stmt.recvDynAck(ctx); err != nil {
		return err
	}

	if err := stmt.recvDoneFinal(ctx); err != nil {
		return err
	}

	return nil
}

// NumInput implements the driver.Stmt interface.
func (stmt Stmt) NumInput() int {
	fieldFmts, err := stmt.fieldFmts()
	if err != nil {
		return -1
	}
	return len(fieldFmts)
}

// Exec implements the driver.Stmt interface.
func (stmt Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), dblib.ValuesToNamedValues(args))
}

// ExecContext implements the driver.StmtExecContext interface.
func (stmt Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	rows, result, err := stmt.GenericExec(ctx, args)
	if rows != nil {
		rows.Close()
	}
	return result, err
}

// Query implements the driver.Stmt interface.
func (stmt Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), dblib.ValuesToNamedValues(args))
}

// QueryContext implements the driver.StmtQueryContext interface.
func (stmt Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := stmt.GenericExec(ctx, args)
	return rows, err
}

// DirectExec is a wrapper for GenericExec and meant to be used when
// directly accessing this library, rather than using database/sql.
//
// The primary advantage are the variadic args, which can be normal
// values and are automatically transformed to driver.NamedValues for
// GenericExec.
func (stmt Stmt) DirectExec(ctx context.Context, args ...interface{}) (driver.Rows, driver.Result, error) {
	var namedArgs []driver.NamedValue
	if len(args) > 0 {
		values := make([]driver.Value, len(args))
		for i, arg := range args {
			values[i] = driver.Value(arg)
		}
		namedArgs = dblib.ValuesToNamedValues(values)
	}

	for i := range args {
		if err := stmt.CheckNamedValue(&namedArgs[i]); err != nil {
			return nil, nil, fmt.Errorf("go-ase: error checking argument: %w", err)
		}
	}

	return stmt.GenericExec(ctx, namedArgs)
}

// GenericExec is the central method through which SQL statements are
// sent to ASE.
func (stmt Stmt) GenericExec(ctx context.Context, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
	// Prepare and send payload
	stmt.pkg.Type = tds.TDS_DYN_EXEC
	if stmt.paramFmt != nil {
		stmt.pkg.Status |= tds.TDS_DYNAMIC_HASARGS
	}
	if err := stmt.conn.Channel.QueuePackage(ctx, stmt.pkg); err != nil {
		return nil, nil, fmt.Errorf("error queueing dynamic statement exec package: %w", err)
	}
	stmt.Reset()

	if stmt.paramFmt != nil {
		if err := stmt.conn.Channel.QueuePackage(ctx, stmt.paramFmt); err != nil {
			return nil, nil, fmt.Errorf("error queueing dynamic statement parameter format: %w", err)
		}

		dataFields := []tds.FieldData{}

		for i, arg := range args {
			fmtField := stmt.paramFmt.Fmts[i]

			dataField, err := tds.LookupFieldData(fmtField)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to find FieldData for datatype %s: %w",
					fmtField.DataType(), err)
			}

			dataField.SetValue(arg.Value)

			dataFields = append(dataFields, dataField)
		}

		if err := stmt.conn.Channel.QueuePackage(ctx, tds.NewParamsPackage(dataFields...)); err != nil {
			return nil, nil, fmt.Errorf("error queueing dynamic statement parameters: %w", err)
		}
	}

	if err := stmt.conn.Channel.SendRemainingPackets(ctx); err != nil {
		return nil, nil, fmt.Errorf("error sending queued packages for dynamic statement execution: %w", err)
	}

	// Receive response
	if err := stmt.recvDynAck(ctx); err != nil {
		return nil, nil, err
	}

	return stmt.conn.genericResults(ctx)
}

// CheckNamedValue implements the driver.NamedValueChecker interface.
func (stmt Stmt) CheckNamedValue(named *driver.NamedValue) error {
	fieldFmts, err := stmt.fieldFmts()
	if err != nil {
		return fmt.Errorf("go-ase: no formats are set: %w", err)
	}

	if named.Ordinal-1 >= len(fieldFmts) {
		return fmt.Errorf("go-ase: ordinal %d (index %d) is larger than the number of expected arguments %d",
			named.Ordinal, named.Ordinal-1, len(fieldFmts))
	}

	val, err := fieldFmts[named.Ordinal-1].DataType().ConvertValue(named.Value)
	if err != nil {
		return fmt.Errorf("go-ase: error converting value: %w", err)
	}

	named.Value = val
	return nil
}
