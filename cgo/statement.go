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
	name        string
	argCount    int
	cmd         *Command
	columnTypes []types.ASEType
}

// Interface satisfaction checks
var (
	_ driver.Stmt             = (*statement)(nil)
	_ driver.StmtExecContext  = (*statement)(nil)
	_ driver.StmtQueryContext = (*statement)(nil)
)

var (
	statementCounter  uint
	statementCounterM = sync.Mutex{}
)

func (conn *Connection) Prepare(query string) (driver.Stmt, error) {
	return conn.prepare(query)
}

func (conn *Connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	recvStmt := make(chan driver.Stmt, 1)
	recvErr := make(chan error, 1)
	go func() {
		stmt, err := conn.prepare(query)
		recvStmt <- stmt
		close(recvStmt)
		recvErr <- err
		close(recvErr)
	}()

	for {
		select {
		case <-ctx.Done():
			go func() {
				stmt := <-recvStmt
				if stmt != nil {
					stmt.Close()
				}
			}()
			return nil, ctx.Err()
		case err := <-recvErr:
			stmt := <-recvStmt
			return stmt, err
		}
	}
}

func (conn *Connection) prepare(query string) (driver.Stmt, error) {
	stmt := &statement{}

	stmt.argCount = strings.Count(query, "?")

	statementCounterM.Lock()
	statementCounter++
	stmt.name = fmt.Sprintf("stmt%d", statementCounter)
	statementCounterM.Unlock()

	cmd, err := conn.Dynamic(stmt.name, query)
	if err != nil {
		stmt.Close()
		return nil, err
	}

	for err = nil; err != io.EOF; _, _, _, err = cmd.Response() {
		if err != nil {
			stmt.Close()
			cmd.Cancel()
			return nil, err
		}
	}

	stmt.cmd = cmd

	err = stmt.fillColumnTypes()
	if err != nil {
		stmt.Close()
		return nil, fmt.Errorf("Failed to retrieve argument types: %v", err)
	}

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
		for err = nil; err != io.EOF; _, _, _, err = stmt.cmd.Response() {
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

		switch stmt.columnTypes[i] {
		case types.IMAGE:
			datafmt.datatype = (C.CS_INT)(types.BINARY)
		default:
			datafmt.datatype = (C.CS_INT)(stmt.columnTypes[i])
		}

		datalen := 0
		var ptr unsafe.Pointer
		switch stmt.columnTypes[i] {
		case types.BIGINT:
			i := (C.CS_BIGINT)(arg.Value.(int64))
			ptr = unsafe.Pointer(&i)
		case types.INT:
			i := (C.CS_INT)(arg.Value.(int32))
			ptr = unsafe.Pointer(&i)
		case types.SMALLINT:
			i := (C.CS_SMALLINT)(arg.Value.(int16))
			ptr = unsafe.Pointer(&i)
		case types.TINYINT:
			i := (C.CS_TINYINT)(arg.Value.(uint8))
			ptr = unsafe.Pointer(&i)
		case types.UBIGINT:
			ci := (C.CS_UBIGINT)(arg.Value.(uint64))
			ptr = unsafe.Pointer(&ci)
		case types.UINT:
			i := (C.CS_UINT)(arg.Value.(uint32))
			ptr = unsafe.Pointer(&i)
		case types.USMALLINT, types.USHORT:
			i := (C.CS_USMALLINT)(arg.Value.(uint16))
			ptr = unsafe.Pointer(&i)
		case types.DECIMAL, types.NUMERIC:
			csDec := (*C.CS_DECIMAL)(C.calloc(1, C.sizeof_CS_DECIMAL))
			defer C.free(unsafe.Pointer(csDec))

			dec := arg.Value.(*types.Decimal)

			csDec.precision = (C.CS_BYTE)(dec.Precision())
			csDec.scale = (C.CS_BYTE)(dec.Scale())

			offset := dec.ByteSize() - len(dec.Bytes())
			for i, b := range dec.Bytes() {
				csDec.array[i+offset] = (C.CS_BYTE)(b)
			}

			if dec.IsNegative() {
				csDec.array[0] = 0x1
			}

			ptr = unsafe.Pointer(csDec)
		case types.FLOAT:
			i := (C.CS_FLOAT)(arg.Value.(float64))
			ptr = unsafe.Pointer(&i)
		case types.REAL:
			i := (C.CS_REAL)(arg.Value.(float64))
			ptr = unsafe.Pointer(&i)
		default:
			return fmt.Errorf("Unhandled column type: %s", stmt.columnTypes[i])
		}

		var csDatalen C.CS_INT
		if datalen != C.CS_UNUSED {
			csDatalen = (C.CS_INT)(datalen)
		} else {
			csDatalen = C.CS_UNUSED
		}

		retval = C.ct_param(stmt.cmd.cmd, datafmt, ptr, csDatalen, 0)
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
	recvErr := make(chan error, 1)
	go func() {
		recvErr <- stmt.exec(args)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-recvErr:
			return err
		}
	}
}

func (stmt *statement) execResults() (driver.Result, error) {
	var resResult driver.Result

	for {
		_, result, _, err := stmt.cmd.Response()
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

func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	err := stmt.exec(libase.ValuesToNamedValues(args))
	if err != nil {
		return nil, err
	}

	return stmt.execResults()
}

func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	err := stmt.execContext(ctx, args)
	if err != nil {
		return nil, err
	}

	recvResult := make(chan driver.Result, 1)
	recvErr := make(chan error, 1)
	go func() {
		res, err := stmt.execResults()
		recvResult <- res
		close(recvResult)
		recvErr <- err
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-recvErr:
			res := <-recvResult
			return res, err
		}
	}
}

func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	err := stmt.exec(libase.ValuesToNamedValues(args))
	if err != nil {
		return nil, err
	}

	rows, _, _, err := stmt.cmd.Response()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	err := stmt.execContext(ctx, args)
	if err != nil {
		return nil, err
	}

	recvRows := make(chan driver.Rows, 1)
	recvErr := make(chan error, 1)
	go func() {
		rows, _, _, err := stmt.cmd.Response()
		recvRows <- rows
		close(recvRows)
		recvErr <- err
		close(recvErr)
	}()

	for {
		select {
		case <-ctx.Done():
			go func() {
				rows := <-recvRows
				if rows != nil {
					rows.Close()
				}
			}()
			return nil, ctx.Err()
		case err := <-recvErr:
			rows := <-recvRows
			return rows, err
		}
	}
}

func (stmt *statement) fillColumnTypes() error {
	name := C.CString(stmt.name)
	defer C.free(unsafe.Pointer(name))

	// Instruct server to send data to descriptor
	retval := C.ct_dynamic(stmt.cmd.cmd, C.CS_DESCRIBE_INPUT, name,
		C.CS_NULLTERM, nil, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Error when preparing input description")
	}

	retval = C.ct_send(stmt.cmd.cmd)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Error sending command to server")
	}

	for {
		_, _, resultType, err := stmt.cmd.Response()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Received error while receiving input description: %v", err)
		}

		if resultType != C.CS_DESCRIBE_RESULT {
			continue
		}

		// Receive number of arguments
		var paramCount C.CS_INT
		retval = C.ct_res_info(stmt.cmd.cmd, C.CS_NUMDATA, unsafe.Pointer(&paramCount), C.CS_UNUSED, nil)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "Failed to retrieve parameter count")
		}

		stmt.argCount = int(paramCount)
		stmt.columnTypes = make([]types.ASEType, stmt.argCount)

		for i := 0; i < stmt.argCount; i++ {
			datafmt := (*C.CS_DATAFMT)(C.calloc(1, C.sizeof_CS_DATAFMT))

			retval = C.ct_describe(stmt.cmd.cmd, (C.CS_INT)(i+1), datafmt)
			if retval != C.CS_SUCCEED {
				return makeError(retval, "Failed to retrieve description of parameter %d", i)
			}

			stmt.columnTypes[i] = types.ASEType(datafmt.datatype)
		}

	}

	return nil
}
