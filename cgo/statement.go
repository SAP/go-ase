package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"

type statement struct {
	query string
	conn  connection
	cmd   *C.CS_COMMAND
}

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

func (statement *statement) Close() error {
	name := C.CString("myquery")
	defer C.free(unsafe.Pointer(name))
	rc := C.ct_dynamic(statement.stmt, C.CS_DEALLOC, name, C.CS_NULLTERM, nil, C.CS_UNUSED)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_dynamic failed")
	}
	return nil
}


func (statement *statement) NumInput() int {
	// TODO
	return -1
}

func (statement *statement) Exec(args []driver.Value) (driver.Result, error) {
	// TODO: bind parameters / args
	// TODO: execute statement
	return &result{stmt: statement.stmt, conn: statement.conn}, nil
}


func (statement *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	value := make([]driver.Value, len(args))
	for i := 0; i < len(args); i++ {
		value[i] = args[i].Value
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return statement.Exec(value)
	} else {
		err := setTimeout(statement, deadline.Sub(time.Now()).Seconds())
		if err != nil {
			return nil, err
		}
		return statement.Exec(value)
	}
}

func (statement *statement) Query(args []driver.Value) (driver.Rows, error) {
	// TODO: bind parameters / args
	// TODO: execute statement
	return &rows{cmd: &csCommand{statement.stmt}}, nil
}

func (statement *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	value := make([]driver.Value, len(args))
	for i := 0; i < len(args); i++ {
		value[i] = args[i].Value
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return statement.Query(value)
	} else {
		err := setTimeout(statement, deadline.Sub(time.Now()).Seconds())
		if err != nil {
			return nil, err
		}
		return statement.Query(value)
	}
}
