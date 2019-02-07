package cgo

//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"fmt"
	"unsafe"
)

// transaction is the struct which represents a database transaction.
type transaction struct {
	conn *connection
}

// Interface satisfaction checks
var _ driver.Tx = (*transaction)(nil)

func (conn *connection) Begin() (driver.Tx, error) {
	return conn.BeginTx(context.Background(), driver.TxOptions{Isolation: 0, ReadOnly: false})
}

func (conn *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if opts.Isolation < 0 || opts.Isolation > 3 {
		return nil, fmt.Errorf("Unsupported isolation level requested: %d", opts.Isolation)
	}

	rc := C.ct_options(conn.conn, C.CS_SET, C.CS_OPT_ISOLATION, unsafe.Pointer(&opts.Isolation), C.CS_UNUSED, nil)
	if rc != C.CS_SUCCEED {
		return nil, makeError(rc, "Failed to set isolation")
	}

	readOnly := C.CS_FALSE
	if opts.ReadOnly {
		readOnly = C.CS_TRUE
	}

	rc = C.ct_con_props(conn.conn, C.CS_SET, C.CS_PROP_READONLY, unsafe.Pointer(&readOnly), C.CS_UNUSED, nil)
	if rc != C.CS_SUCCEED {
		return nil, makeError(rc, "Failed to set readonly")
	}

	// TODO disable autocommit

	return &transaction{conn: conn}, nil
}

func (transaction *transaction) Commit() error {
	// TODO
	return nil
}

func (transaction *transaction) Rollback() error {
	// TODO
	return nil
}
