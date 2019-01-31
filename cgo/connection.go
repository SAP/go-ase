package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
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
	rows, err := conn.QueryContext(ctx, "SELECT 'PING'", []driver.NamedValue{})
	if err != nil {
		return driver.ErrBadConn
	}
	defer rows.Close()

	cols := rows.Columns()
	cellRefs := make([]driver.Value, len(cols))

	for {
		err := rows.Next(cellRefs)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Error occurred while exhausting result set: %v", err)
		}
	}

	return nil
}

func (conn *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	return conn.ExecContext(context.Background(), query, libase.ValuesToNamedValues(args))
}

func (conn *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	cmd, err := conn.execContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}
	defer cmd.drop()

	var resResult driver.Result

	for rows, result, err := cmd.resultsHelper(); err != io.EOF; rows, result, err = cmd.resultsHelper() {
		if err != nil {
			return nil, fmt.Errorf("Received error reading results: %v", err)
		}

		if rows != nil {
			log.Printf("rows is not nil - this should not be the case")
		}

		if result != nil {
			if resResult != nil {
				return nil, fmt.Errorf("Received more than one result: %v, %v", resResult, result)
			}
			resResult = result
		}
	}

	return resResult, nil
}

func (conn *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	return conn.QueryContext(context.Background(), query, libase.ValuesToNamedValues(args))
}

func (conn *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cmd, err := conn.execContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}

	rows, result, err := cmd.resultsHelper()
	if err != nil {
		return nil, fmt.Errorf("Received error while retrieving results: %v", err)
	}

	if result != nil {
		return nil, fmt.Errorf("Received result when querying: %v", result)
	}

	return rows, nil
}
