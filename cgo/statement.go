package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"unsafe"

	"github.com/SAP/go-ase/libase"
	"github.com/SAP/go-ase/libase/types"
)

type statement struct {
	name     string
	argCount int
	cmd      *csCommand
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

	stmt.argCount = strings.Count(query, "?")

	statementCounterM.Lock()
	statementCounter += 1
	stmt.name = fmt.Sprintf("stmt%d", statementCounter)
	statementCounterM.Unlock()

	cmd, err := conn.dynamic(stmt.name, query)
	if err != nil {
		stmt.Close()
		return nil, err
	}

	for err = nil; err != io.EOF; _, _, err = cmd.resultsHelper() {
		if err != nil {
			stmt.Close()
			cmd.cancel()
			return nil, err
		}
	}

	stmt.cmd = cmd

	return stmt, nil
}

func (stmt *statement) Close() error {
	if stmt.cmd != nil {
		name := C.CString(stmt.name)
		defer C.free(unsafe.Pointer(name))

		retval := C.ct_dynamic(stmt.cmd.cmd, C.CS_DEALLOC, name, C.CS_NULLTERM, nil, C.CS_UNUSED)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_dynamic with C.CS_DEALLOC failed")
		}

		retval = C.ct_send(stmt.cmd.cmd)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_send failed")
		}

		var err error
		for err = nil; err != io.EOF; _, _, err = stmt.cmd.resultsHelper() {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (stmt *statement) NumInput() int {
	return stmt.argCount
}

func (stmt *statement) exec(args []driver.NamedValue) (driver.Rows, driver.Result, error) {
	if len(args) != stmt.argCount {
		return nil, nil, fmt.Errorf("Mismatched argument count - expected %d, got %d",
			stmt.argCount, len(args))
	}

	name := C.CString(stmt.name)
	defer C.free(unsafe.Pointer(name))

	retval := C.ct_dynamic(stmt.cmd.cmd, C.CS_EXECUTE, name, C.CS_NULLTERM, nil, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_dynamic with CS_EXECUTE failed")
	}

	for i, arg := range args {
		datafmt := (*C.CS_DATAFMT)(C.calloc(1, C.sizeof_CS_DATAFMT))
		defer C.free(unsafe.Pointer(datafmt))
		datafmt.status = C.CS_INPUTVALUE
		datafmt.namelen = C.CS_NULLTERM
		asetype, err := types.FromGoType(arg.Value)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to retrieve ASEType for driver.Value: %v", err)
		}
		datafmt.datatype = (C.CS_INT)(asetype)

		switch arg.Value.(type) {
		case []byte:
			datafmt.format = C.CS_FMT_PADNULL
		case string:
			datafmt.format = C.CS_FMT_NULLTERM
			datafmt.maxlength = C.CS_MAX_CHAR
		}

		s := fmt.Sprintf("%v", arg.Value)
		datalen := len(s)
		ptr := C.CString(s)
		defer C.free(unsafe.Pointer(ptr))

		retval = C.ct_param(stmt.cmd.cmd, datafmt, unsafe.Pointer(ptr), (C.CS_INT)(datalen), 0)
		if retval != C.CS_SUCCEED {
			return nil, nil, makeError(retval, "C.ct_param on parameter %d failed with argument '%v'", i, arg)
		}
	}

	retval := C.ct_send(stmt.cmd.cmd)
	if retval != C.CS_SUCCEED {
		return nil, nil, makeError(retval, "C.ct_Send failed")
	}

	return stmt.cmd.resultsHelper()
}

func (stmt *statement) execContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, driver.Result, error) {
	return stmt.exec(args)
}

func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	_, result, err := stmt.execContext(ctx, args)
	return result, err
}

func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := stmt.execContext(ctx, args)
	return rows, err
}
