// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql/driver"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/SAP/go-ase/libase"
	"github.com/SAP/go-ase/libase/tds"
)

var (
	_ driver.Stmt              = (*Stmt)(nil)
	_ driver.StmtExecContext   = (*Stmt)(nil)
	_ driver.StmtQueryContext  = (*Stmt)(nil)
	_ driver.NamedValueChecker = (*Stmt)(nil)

	stmtIdCounter uint64 = 0
	stmtIdPool           = &sync.Pool{
		New: func() interface{} {
			newId := atomic.AddUint64(&stmtIdCounter, 1)
			return &newId
		},
	}
)

type Stmt struct {
	conn *Conn

	stmtId *uint64
	pkg    *tds.DynamicPackage

	paramFmt *tds.ParamFmtPackage
	rowFmt   *tds.RowFmtPackage
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	// TODO option for create_proc
	return c.NewStmt(ctx, "", query, true)
}

func (c *Conn) NewStmt(ctx context.Context, name, query string, create_proc bool) (*Stmt, error) {
	stmt := &Stmt{conn: c}

	if name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("go-ase: no name for dynamic SQL passed and hostname cannot be retrieved: %w", err)
		}

		hostname = strings.Replace(hostname, "-", "_", -1)

		// TODO different pools for procs and prepares
		stmt.stmtId = stmtIdPool.Get().(*uint64)
		name = fmt.Sprintf("goase_%s_%d_%d", hostname, os.Getpid(), *stmt.stmtId)
	}

	stmt.pkg = &tds.DynamicPackage{
		ID: name,
	}

	if create_proc {
		stmt.pkg.Stmt = fmt.Sprintf("create proc %s as %s", name, query)
	} else {
		stmt.pkg.Stmt = query
	}

	// Reset statement to default before proceeding
	stmt.Reset()

	err := stmt.allocateOnServer(ctx)
	if err != nil {
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

	err := stmt.recvDynAck(ctx)
	if err != nil {
		return err
	}

	_, err = stmt.conn.Channel.NextPackageUntil(ctx, true,
		func(pkg tds.Package) (bool, error) {
			switch typed := pkg.(type) {
			case *tds.ParamFmtPackage:
				stmt.paramFmt = typed
				return false, nil
			case *tds.RowFmtPackage:
				stmt.rowFmt = typed
				return false, nil
			case *tds.DonePackage:
				if typed.Status != tds.TDS_DONE_FINAL {
					return false, fmt.Errorf("DonePackage does not have status TDS_DONE_FINAL set: %s", typed)
				}
				return true, nil
			default:
				return false, fmt.Errorf("unexpected package received: %#v", typed)
			}
		},
	)
	if err != nil {
		stmt.close(ctx)
		return err
	}

	return nil
}

func (stmt *Stmt) Reset() {
	stmt.pkg.Type = tds.TDS_DYN_INVALID
	stmt.pkg.Status = tds.TDS_DYNAMIC_UNUSED
}

func (stmt *Stmt) Close() error {
	return stmt.close(context.Background())
}

func (stmt *Stmt) close(ctx context.Context) error {
	if stmt.stmtId != nil {
		defer stmtIdPool.Put(stmt.stmtId)
	}

	// communicate deallocation with server
	// TODO option to not deallocate procs
	stmt.pkg.Type = tds.TDS_DYN_DEALLOC
	err := stmt.conn.Channel.SendPackage(ctx, stmt.pkg)
	if err != nil {
		return fmt.Errorf("error sending dealloc package: %w", err)
	}
	stmt.Reset()

	err = stmt.recvDynAck(ctx)
	if err != nil {
		return err
	}

	err = stmt.recvDoneFinal(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (stmt Stmt) NumInput() int {
	return len(stmt.paramFmt.Fmts)
}

func (stmt Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	rows, result, err := stmt.exec(ctx, args)
	if rows != nil {
		rows.Close()
	}
	return result, err
}

func (stmt Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := stmt.exec(ctx, args)
	return rows, err
}

func (stmt Stmt) exec(ctx context.Context, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
	// Prepare and send payload
	stmt.pkg.Type = tds.TDS_DYN_EXEC
	if stmt.paramFmt != nil {
		stmt.pkg.Status |= tds.TDS_DYNAMIC_HASARGS
	}
	err := stmt.conn.Channel.QueuePackage(ctx, stmt.pkg)
	if err != nil {
		return nil, nil, fmt.Errorf("error queueing dynamic statement exec package: %w", err)
	}
	stmt.Reset()

	if stmt.paramFmt != nil {
		err := stmt.conn.Channel.QueuePackage(ctx, stmt.paramFmt)
		if err != nil {
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
	err = stmt.recvDynAck(ctx)
	if err != nil {
		return nil, nil, err
	}

	return stmt.conn.genericResults(ctx)
}

func (stmt Stmt) CheckNamedValues(namedValues []*driver.NamedValue) error {
	for _, nv := range namedValues {
		err := stmt.CheckNamedValue(nv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (stmt Stmt) CheckNamedValue(named *driver.NamedValue) error {
	var fieldFmts []tds.FieldFmt
	if stmt.paramFmt != nil {
		fieldFmts = stmt.paramFmt.Fmts
	} else if stmt.rowFmt != nil {
		fieldFmts = stmt.rowFmt.Fmts
	} else {
		return fmt.Errorf("go-ase: both row and paramFmt are unset")
	}

	index := named.Ordinal - 1
	if index > len(fieldFmts) {
		return fmt.Errorf("go-ase: ordinal %d is larger than the number of columns %d",
			named.Ordinal, len(fieldFmts))
	}

	val, err := fieldFmts[index].DataType().ConvertValue(named.Value)
	if err != nil {
		return fmt.Errorf("go-ase: error converting value: %w", err)
	}

	named.Value = val
	return nil
}
