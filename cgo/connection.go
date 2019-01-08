package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/SAP/go-ase/libase"
)

// connection is the struct which represents a database connection.
type connection struct {
	conn *C.CS_CONNECTION
}

// newConnection allocated initializes a new connection based on the
// options in the dsn.
func newConnection(dsn libase.DsnInfo) (*connection, error) {
	err := driverCtx.init()
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure context: %v", err)
	}

	conn := &connection{}

	retval := C.ct_con_alloc(driverCtx.ctx, &conn.conn)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_alloc failed")
	}

	// Set username.
	username := unsafe.Pointer(C.CString(dsn.Username))
	defer C.free(username)
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_USERNAME, username, C.CS_NULLTERM, nil)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_USERNAME")
	}

	// Set password encryption
	cTrue := C.CS_TRUE
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SEC_EXTENDED_ENCRYPTION,
		unsafe.Pointer(&cTrue), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SEC_EXTENDED_ENCRYPTION")
	}

	cFalse := C.CS_FALSE
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SEC_NON_ENCRYPTION_RETRY,
		unsafe.Pointer(&cFalse), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SEC_NON_ENCRYPTION_RETRY")
	}

	// Set password.
	password := unsafe.Pointer(C.CString(dsn.Password))
	defer C.free(password)
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_PASSWORD, password, C.CS_NULLTERM, nil)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_PASSWORD")
	}

	// Set hostname and port.
	hostport := unsafe.Pointer(C.CString(dsn.Host + " " + dsn.Port))
	defer C.free(hostport)
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SERVERADDR, hostport, C.CS_NULLTERM, nil)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SERVERADDR")
	}

	retval = C.ct_connect(conn.conn, nil, 0)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_connect failed")
	}

	// // Set database
	// if dsn.Database != "" {
	// 	database := unsafe.Pointer(C.CString(dsn.Database))
	// 	defer C.free(database)
	// 	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_PROP_INITIAL_DATABASE, database,
	// 		C.CS_NULLTERM, nil)
	// 	if retval != C.CS_SUCCEED {
	// 		conn.Close()
	// 		return nil, makeError(retval, "C.ct_con_props failed for CS_PROP_INITIAL_DATABASE")
	// 	}
	// }

	return conn, nil
}

func (conn *connection) ResetSession(ctx context.Context) error {
	// TODO
	return nil
}

// Close closes and deallocates a connection.
func (conn *connection) Close() error {
	// Call context.drop when exiting this function to decrease the
	// connection counter and potentially deallocate the context.
	defer driverCtx.drop()

	retval := C.ct_close(conn.conn, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_close failed, connection has results pending")
	}

	retval = C.ct_con_drop(conn.conn)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_con_drop failed")
	}

	conn.conn = nil
	return nil
}

func (conn *connection) Ping(ctx context.Context) error {
	//TODO context
	cmd, err := conn.exec("SELECT 'PING'")
	if err != nil {
		return driver.ErrBadConn
	}

	_, _, err = cmd.results()
	if err != nil {
		return driver.ErrBadConn
	}

	err = cmd.cancel()
	if err != nil {
		return driver.ErrBadConn
	}

	return nil
}


func (conn *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	// TODO: driver.Value handling

	cmd, err := conn.exec(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}
	defer cmd.drop()

	rows, result, err := cmd.results()
	if err != nil {
		return nil, fmt.Errorf("Received error when reading results: %v", err)
	}

	if rows != nil {
		return nil, fmt.Errorf("Received rows when executing an exec")
	}

	return result, nil
}

func (conn *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	// TODO
	return nil, nil
}

func (conn *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	cmd, err := conn.exec(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}

	rows, result, err := cmd.results()
	if err != nil {
		return nil, fmt.Errorf("Received error when preparing rows: %v", err)
	}

	if result != nil {
		return nil, fmt.Errorf("Received results when querying")
	}

	return rows, nil
}

func (conn *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	//TODO
	return nil, nil
}

func (conn *connection) Prepare(query string) (driver.Stmt, error) {
	psql := C.CString(query)
	defer C.free(unsafe.Pointer(psql))
	name := C.CString("myquery")
	defer C.free(unsafe.Pointer(name))
	var cPreparedStatement *C.CS_COMMAND
	rc := C.ct_dynamic(cPreparedStatement, C.CS_PREPARE, name, C.CS_NULLTERM, psql, C.CS_NULLTERM)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_dynamic failed")
	}
	return &statement{query: query, stmt: cPreparedStatement, conn: conn.conn}, nil
}

func (conn *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	//TODO
	return nil, nil
}

func (conn *connection) Begin() (driver.Tx, error) {
	return conn.BeginTx(context.Background(), driver.TxOptions{Isolation: 0, ReadOnly: false})
}


func (conn *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if opts.IsolationLevel < 0 || opts.IsolationLevel > 3 {
		return nil, fmt.Errorf("Unsupported isolation level requested: %d", isolationLevel)
	}

	rc := C.ct_options(connection.conn, C.CS_SET, C.CS_OPT_ISOLATION, unsafe.Pointer(&opts.IsolationLevel), C.CS_UNUSED, nil)
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
