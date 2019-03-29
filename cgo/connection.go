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
	"github.com/SAP/go-ase/libase/dsn"
)

// connection is the struct which represents a database connection.
type connection struct {
	conn      *C.CS_CONNECTION
	driverCtx *csContext
}

// Interface satisfaction checks
var (
	_ driver.Conn               = (*connection)(nil)
	_ driver.ConnBeginTx        = (*connection)(nil)
	_ driver.ConnPrepareContext = (*connection)(nil)
	_ driver.Execer             = (*connection)(nil)
	_ driver.ExecerContext      = (*connection)(nil)
	_ driver.Pinger             = (*connection)(nil)
	_ driver.Queryer            = (*connection)(nil)
	_ driver.QueryerContext     = (*connection)(nil)
)

// newConnection allocated initializes a new connection based on the
// options in the dsn.
//
// If driverCtx is nil a new csContext will be initialized.
func newConnection(driverCtx *csContext, dsn dsn.DsnInfo) (*connection, error) {
	if driverCtx == nil {
		var err error
		driverCtx, err = newCsContext(dsn)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize context for conn: %v", err)
		}
	}

	err := driverCtx.newConn()
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure context: %v", err)
	}

	conn := &connection{
		driverCtx: driverCtx,
	}

	retval := C.ct_con_alloc(driverCtx.ctx, &conn.conn)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_alloc failed")
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

	// Give preference to the user store key
	if len(dsn.Userstorekey) > 0 {
		// Set userstorekey
		userstorekey := unsafe.Pointer(C.CString(dsn.Userstorekey))
		defer C.free(userstorekey)
		retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SECSTOREKEY, userstorekey, C.CS_NULLTERM, nil)
		if retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for C.CS_SECSTOREKEY")
		}
	} else {
		// Set username.
		username := unsafe.Pointer(C.CString(dsn.Username))
		defer C.free(username)
		retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_USERNAME, username, C.CS_NULLTERM, nil)
		if retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for CS_USERNAME")
		}

		// Set password.
		password := unsafe.Pointer(C.CString(dsn.Password))
		defer C.free(password)
		retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_PASSWORD, password, C.CS_NULLTERM, nil)
		if retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for CS_PASSWORD")
		}
	}

	if dsn.Host != "" && dsn.Port != "" {
		// Set hostname and port.
		hostport := unsafe.Pointer(C.CString(dsn.Host + " " + dsn.Port))
		defer C.free(hostport)
		retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SERVERADDR, hostport, C.CS_NULLTERM, nil)
		if retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for CS_SERVERADDR")
		}
	}

	retval = C.ct_connect(conn.conn, nil, 0)
	if retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_connect failed")
	}

	// Set database
	if dsn.Database != "" {
		_, err := conn.Exec("use "+dsn.Database, nil)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("Failed to connect to database %s: %v", dsn.Database, err)
		}
	}

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
	defer conn.driverCtx.dropConn()

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
	q, err := libase.QueryFormat(query, args...)
	if err != nil {
		return nil, err
	}

	return conn.ExecContext(context.Background(), q, nil)
}

func (conn *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	q, err := libase.NamedQueryFormat(query, args...)
	if err != nil {
		return nil, err
	}

	cmd, err := conn.execContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}
	defer cmd.drop()

	var result, resResult driver.Result
	for _, result, err = cmd.resultsHelper(); err != io.EOF; _, result, err = cmd.resultsHelper() {
		if err != nil {
			return nil, fmt.Errorf("Received error reading results: %v", err)
		}

		if result != nil {
			resResult = result
		}
	}

	return resResult, nil
}

func (conn *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	q, err := libase.QueryFormat(query, args...)
	if err != nil {
		return nil, err
	}

	return conn.QueryContext(context.Background(), q, nil)
}

func (conn *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q, err := libase.NamedQueryFormat(query, args...)
	if err != nil {
		return nil, err
	}

	cmd, err := conn.execContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("Failed to send command: %v", err)
	}

	rows, _, err := cmd.resultsHelper()
	if err != nil {
		return nil, fmt.Errorf("Received error while retrieving results: %v", err)
	}

	return rows, nil
}
