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

func (stmt *statement) exec(args []driver.NamedValue) error {
	if len(args) != stmt.argCount {
		return fmt.Errorf("Mismatched argument count - expected %d, got %d",
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
			return fmt.Errorf("Failed to retrieve ASEType for driver.Value: %v", err)
		}
		datafmt.datatype = (C.CS_INT)(asetype)

		datalen := 0
		var ptr unsafe.Pointer
		switch arg.Value.(type) {
		case int64:
			i := (C.CS_BIGINT)(arg.Value.(int64))
			ptr = unsafe.Pointer(&i)
		case uint64:
			i := (C.CS_UBIGINT)(arg.Value.(uint64))
			ptr = unsafe.Pointer(&i)
		case float64:
			i := (C.CS_FLOAT)(arg.Value.(float64))
			ptr = unsafe.Pointer(&i)
		case bool:
			b := (C.CS_BOOL)(0)
			if arg.Value.(bool) {
				b = (C.CS_BOOL)(1)
			}
			ptr = unsafe.Pointer(&b)
			datalen = 1
		case []byte:
			ptr = C.CBytes(arg.Value.([]byte))
			defer C.free(ptr)
			datafmt.format = C.CS_FMT_PADNULL
			datalen = len(arg.Value.([]byte))
		case string:
			ptr = unsafe.Pointer(C.CString(arg.Value.(string)))
			defer C.free(ptr)
			datafmt.format = C.CS_FMT_NULLTERM
			datafmt.maxlength = C.CS_MAX_CHAR
			datalen = len(arg.Value.(string))
		default:
			return fmt.Errorf("Unable to transform to Client-Library: %v", arg.Value)
		}

		retval = C.ct_param(stmt.cmd.cmd, datafmt, ptr, (C.CS_INT)(datalen), 0)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_param on parameter %d failed with argument '%v'", i, arg)
		}
	}

	retval = C.ct_send(stmt.cmd.cmd)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_send failed")
	}

	return nil
}

func (stmt *statement) execContext(ctx context.Context, args []driver.NamedValue) error {
	return stmt.exec(args)
}

func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	err := stmt.execContext(ctx, args)
	if err != nil {
		return nil, err
	}

	var resResult driver.Result

	for {
		_, result, err := stmt.cmd.resultsHelper()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if result != nil {
			resResult = result
		}
	}

	return resResult, nil
}

func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), libase.ValuesToNamedValues(args))
}

func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	err := stmt.execContext(ctx, args)
	if err != nil {
		return nil, err
	}

	rows, _, err := stmt.cmd.resultsHelper()
	if err != nil {
		return nil, err
	}

	return rows, nil
}
