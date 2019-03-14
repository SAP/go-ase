package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"errors"
	"sync"
	"unsafe"

	"github.com/SAP/go-ase/libase"
)

type statement struct {
	name  *C.char
	query string
	cmd   *C.CS_COMMAND
}

// Interface satisfaction checks
var (
	_ driver.Stmt             = (*statement)(nil)
	_ driver.StmtExecContext  = (*statement)(nil)
	_ driver.StmtQueryContext = (*statement)(nil)
)

var (
	statementCounter  uint = 0
	statementCounterM      = sync.Mutex{}
)

func (conn *connection) Prepare(query string) (driver.Stmt, error) {
	return conn.PrepareContext(context.Background(), query)
}

func (conn *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	stmt := &statement{}

	drv.statementCounterM.Lock()
	drv.statementCounter += 1
	stmt.name = C.CString(string(drv.statementCounter))
	drv.statementCounterM.Unlock()

	q := C.CString(query)
	defer C.free(unsafe.Pointer(q))
	retval := C.ct_dynamic(stmt.cmd, C.CS_PREPARE, stmt.name, C.CS_NULLTERM, q, C.CS_NULLTERM)
	if retval != C.CS_SUCCEED {
		stmt.Close()
		return nil, makeError(retval, "Failed to initialize dynamic command")
	}

	return stmt, nil
}

func (stmt *statement) Close() error {

	rc := C.ct_dynamic(stmt.cmd, C.CS_DEALLOC, stmt.name, C.CS_NULLTERM, nil, C.CS_UNUSED)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_dynamic failed")
	}

	C.free(unsafe.Pointer(stmt.name))

	return nil
}

func (stmt *statement) NumInput() int {
	// TODO
	return -1
}

func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	// TODO
	return nil, nil
}

func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	// TODO
	return nil, nil
}
