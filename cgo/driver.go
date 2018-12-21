package cgo

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -Wl,-rpath,\$ORIGIN/../lib
#include <stdlib.h>
#include "ctlib.h"
#include "bridge.h"
*/
import "C"
import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/SAP/go-ase/libase"
	"github.com/pkg/errors"
)

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

// drv is the struct on which we later call Open() to get a connection.
type drv struct{}

// connWrapper is a helper struct as we cannot pass pointers to pointers with cgo.
// For this we have bridge.c and bridge.h which implement a C struct CS_CONNECTION_WRAPPER
// and a C function ct_con_alloc_wrapper() to wrap the C function ct_con_alloc() which
// expects a pointer to a CS_CONNECTION pointer.
type connWrapper struct {
	conn *C.CS_CONNECTION
	rc   C.CS_RETCODE
}

// transaction is the struct which represents a database transaction.
type transaction struct {
	conn *C.CS_CONNECTION
}

// statement is the struct which represents a database statement
type statement struct {
	query string
	stmt  *C.CS_COMMAND
	conn  *C.CS_CONNECTION
	Ok    bool
}

// rows is the struct which represents a database result set
type rows struct {
	stmt *C.CS_COMMAND
	conn *C.CS_CONNECTION
}

//keep track of rows affected after inserts and updates
type result struct {
	stmt         *C.CS_COMMAND
	conn         *C.CS_CONNECTION
	rowsAffected int64
}

func init() {
	sql.Register(DriverName, &drv{})
}

func (d *drv) Open(dsn string) (driver.Conn, error) {
	dsnInfo, err := libase.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN: %v", err)
	}

	return newConnection(*dsnInfo)
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
	return connection.drop()
}

func (connection *connection) Begin() (driver.Tx, error) {
	return connection.BeginTx(context.Background(), driver.TxOptions{Isolation: 0, ReadOnly: false})
}

func (connection *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	isolationLevel := int(opts.Isolation)
	if isolationLevel < 0 || isolationLevel > 3 {
		return nil, fmt.Errorf("Unsupported isolation level requested: %d", isolationLevel)
	}

	rc := C.ct_options(connection.conn, C.CS_SET, C.CS_OPT_ISOLATION, unsafe.Pointer(&isolationLevel), C.CS_UNUSED, nil)
	if rc != C.CS_SUCCEED {
		return nil, makeError(rc, "Failed to set isolation")
	}

	readOnly := C.CS_FALSE
	if opts.ReadOnly {
		readOnly = C.CS_TRUE
	}

	rc = C.ct_con_props(connection.conn, C.CS_SET, C.CS_PROP_READONLY, unsafe.Pointer(&readOnly), C.CS_UNUSED, nil)
	if rc != C.CS_SUCCEED {
		return nil, makeError(rc, "Failed to set readonly")
	}

	// TODO disable autocommit

	return &transaction{conn: connection.conn}, nil
}

func (conn *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	// TODO: driver.Value handling

	cmd, cmdFree, err := conn.exec(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}
	defer cmdFree()

	rows, result, err := results(cmd)
	if err != nil {
		return nil, fmt.Errorf("Received error when reading results: %v", err)
	}

	if rows != nil {
		return nil, fmt.Errorf("Received rows when executing a language command")
	}

	return result, nil
}

func (connection *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	// TODO
	return nil, nil
}

func (rows *rows) Close() error {
	// TODO
	return nil
}

func (rows *rows) Columns() []string {
	// TODO
	columnNames := make([]string, 1)
	return columnNames
}

func (rows *rows) ColumnTypeDatabaseTypeName(index int) string {
	// TODO
	return ""
}

func (rows *rows) ColumnTypeNullable(index int) (bool, bool) {
	// TODO
	return false, false
}

func (rows *rows) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	// TODO
	return 0, 0, false
}

func (rows *rows) ColumnTypeLength(index int) (int64, bool) {
	// TODO
	return 0, false
}

func (rows *rows) ColumnTypeScanType(index int) reflect.Type {
	// TODO
	return nil
}

func (rows *rows) Next(dest []driver.Value) error {
	// TODO
	return nil
}

func (rows *rows) HasNextResultSet() bool {
	return true
}

func (rows *rows) NextResultSet() error {
	// TODO
	return nil
}

func (connection *connection) Ping(ctx context.Context) error {
	var cCmd *C.CS_COMMAND

	// allocate the command
	rc := C.ct_cmd_alloc(connection.conn, &cCmd)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_cmd_alloc failed")
	}
	// at the end drop the command
	defer C.ct_cmd_drop(cCmd)

	// fill the command
	cQuery := C.CString("SELECT 'PING'")
	rc = C.ct_command(cCmd, C.CS_LANG_CMD, cQuery, C.CS_NULLTERM, C.CS_UNUSED)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_command failed")
	}

	// send the command
	rc = C.ct_send(cCmd)
	if rc != C.CS_SUCCEED {
		return driver.ErrBadConn
	}

	// cancel the results
	rc = C.ct_cancel(nil, cCmd, C.CS_CANCEL_ALL)
	if rc != C.CS_SUCCEED {
		return errors.New("C.ct_cancel failed")
	}

	return nil
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

func setTimeout(statement *statement, timeout float64) error {
	// TODO
	return nil
}

func (statement *statement) Query(args []driver.Value) (driver.Rows, error) {
	// TODO: bind parameters / args
	// TODO: execute statement
	return &rows{stmt: statement.stmt, conn: statement.conn}, nil
}

func (result result) LastInsertId() (int64, error) {
	return -1, errors.New("Feature not supported")
}

func (result result) RowsAffected() (int64, error) {
	if result.rowsAffected == -1 {
		return -1, errors.New("Value unset")
	}
	return result.rowsAffected, nil
}

func (statement *statement) bindParameter(index int, paramVal driver.Value) error {
	// TODO
	return nil
}

func (transaction *transaction) Commit() error {
	// TODO
	return nil
}

func (transaction *transaction) Rollback() error {
	// TODO
	return nil
}

// needed to handle nil time values
type NullTime struct {
	Time  time.Time
	Valid bool
}

// needed to handle nil binary values
type NullBytes struct {
	Bytes []byte
	Valid bool
}

func (nullTime *NullTime) Scan(value interface{}) error {
	nullTime.Time, nullTime.Valid = value.(time.Time)
	return nil
}

func (nullTime NullTime) Value() (driver.Value, error) {
	if !nullTime.Valid {
		return nil, nil
	}
	return nullTime.Time, nil
}

func (nullBytes *NullBytes) Scan(value interface{}) error {
	nullBytes.Bytes, nullBytes.Valid = value.([]byte)
	return nil
}

func (nullBytes NullBytes) Value() (driver.Value, error) {
	if !nullBytes.Valid {
		return nil, nil
	}
	return nullBytes.Bytes, nil
}
