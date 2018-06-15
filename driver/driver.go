package driver

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -Wl,-rpath,\$ORIGIN
#include <stdlib.h>
#include "ctpublic.h"
*/
import "C"
import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
	"unsafe"
)

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

type drv struct{}

var (
	cContext *C.CS_CONTEXT
)

//database connection
type connection struct {
	conn *C.CS_CONNECTION
}

type transaction struct {
	conn *C.CS_CONNECTION
}

//statement
type statement struct {
	query string
	stmt  *C.CS_COMMAND
	conn  *C.CS_CONNECTION
	Ok    bool
}

//result set
type rows struct {
	stmt *C.CS_COMMAND
	conn *C.CS_CONNECTION
}

//keep track of rows affected after inserts and updates
type result struct {
	stmt *C.CS_COMMAND
	conn *C.CS_CONNECTION
}

//needed to handle nil time values
type NullTime struct {
	Time  time.Time
	Valid bool
}

//needed to handle nil binary values
type NullBytes struct {
	Bytes []byte
	Valid bool
}

func init() {
	sql.Register(DriverName, &drv{})
	rc := C.cs_ctx_alloc(C.CS_CURRENT_VERSION, &cContext)
	if rc != C.CS_SUCCEED {
		fmt.Println("C.cs_ctx_alloc failed")
		return
	}
	defer C.free(unsafe.Pointer(cContext))

	rc = C.ct_init(cContext, C.CS_CURRENT_VERSION)
	if rc != C.CS_SUCCEED {
		fmt.Println("C.ct_init failed")
		C.cs_ctx_drop(cContext)
		return
	}
}

func (d *drv) Open(dsn string) (driver.Conn, error) {
	dsnInfo, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// create connection
	var cConnection *C.CS_CONNECTION
	rc := C.ct_con_alloc(cContext, &cConnection)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_con_alloc failed")
	}

	// set user name
	cUsername := unsafe.Pointer(&dsnInfo.Username)
	defer C.free(unsafe.Pointer(cUsername))
	rc = C.ct_con_props(cConnection, C.CS_SET, C.CS_USERNAME, cUsername, C.CS_NULLTERM, nil)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_con_props failed for C.CS_USERNAME")
	}

	// set password
	cPassword := unsafe.Pointer(&dsnInfo.Password)
	defer C.free(unsafe.Pointer(cPassword))
	rc = C.ct_con_props(cConnection, C.CS_SET, C.CS_PASSWORD, cPassword, C.CS_NULLTERM, nil)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_con_props failed for C.CS_PASSWORD")
	}

	// connect
	cHostname := C.CString(dsnInfo.Host)
	cNullterm := (C.long)(C.CS_NULLTERM)
	if dsnInfo.Host != "" {
		cNullterm = (C.long)(0)
	}
	rc = C.ct_connect(cConnection, cHostname, cNullterm)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_connect failed")
	}

	// return connection
	return &connection{conn: cConnection}, nil
}

func (connection *connection) Prepare(query string) (driver.Stmt, error) {
	psql := C.CString(query)
	defer C.free(unsafe.Pointer(psql))
	name := C.CString("myquery")
	defer C.free(unsafe.Pointer(name))
	var cPreparedStatement *C.CS_COMMAND
	rc := C.ct_dynamic(cPreparedStatement, C.CS_PREPARE, name, C.CS_NULLTERM, psql, C.CS_NULLTERM)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_dynamic failed")
	}
	return &statement{query: query, stmt: cPreparedStatement, conn: connection.conn}, nil
}

func (connection *connection) Close() error {
	rc := C.ct_close(connection.conn, C.CS_UNUSED)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_close failed")
	}
	rc = C.ct_con_drop(connection.conn)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_con_drop failed")
	}
	return nil
}

func (connection *connection) Begin() (driver.Tx, error) {
	return &transaction{conn: connection.conn}, nil
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

func (statement *statement) Query(args []driver.Value) (driver.Rows, error) {
	// TODO: bind parameters / args
	// TODO: execute statement
	return &rows{stmt: statement.stmt, conn: statement.conn}, nil
}

func (transaction *transaction) Rollback() error {
	// TODO
	return nil
}

func (transaction *transaction) Commit() error {
	// TODO
	return nil
}
