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
	// ASE does not support read-only transactions - the connection
	// itself however can be set as read-only.
	// readonlyPreTx signals if a connection was marked as read-only
	// before the transaction startet.
	readonlyPreTx bool
	// readonlyNeedsReset is true if the read-only option passed to
	// BeginTx differs from the read-only property of the connection.
	readonlyNeedsReset bool
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

	tx := &transaction{conn, false, false}

	_, err := tx.conn.Exec("BEGIN TRANSACTION", nil)
	if err != nil {
		return fmt.Errorf("Failed to start transaction: %v", err)
	}

	_, err := tx.conn.Exec("SET TRANSACTION ISOLATION LEVEL ?", []driver.Value{opts.Isolation})
	if err != nil {
		return fmt.Errorf("Failed to set isolation level for transaction: %v", err)
	}

	ro := C.CS_FALSE
	rc = C.ct_con_preops(tx.conn.conn, C.CS_GET, C.CS_PROP_READONLY, unsafe.Pointer(&ro), C.CS_UNUSED, nil)
	if rc != C.CS_SUCCEED {
		return nil, makeError(rc, "Failed to retrieve readonly property")
	}

	if opts.Readonly != bool(ro) {
		tx.readonlyPreTx = bool(ro)
		tx.readonlyNeedsReset = true

		err = tx.setRO(opts.ReadOnly)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (tx *transaction) Commit() error {
	_, err := tx.conn.Exec("COMMIT TRANSACTION", nil)
	if err != nil {
		return err
	}
	return tx.finish()
}

func (tx *transaction) Rollback() error {
	_, err := tx.conn.Exec("ROLLBACK TRANSACTION", nil)
	if err != nil {
		return err
	}
	return tx.finish()
}

func (tx *transaction) finish() error {
	if tx.readonlyNeedsReset {
		err := tx.setRO(tx.roPreTx)
		if err != nil {
			return err
		}
	}
	tx.conn = nil
	return nil
}

func (tx *transaction) setRO(ro bool) error {
	ro := C.CS_FALSE
	if ro {
		ro = C.CS_TRUE
	}

	rc = C.ct_con_props(tx.conn.conn, C.CS_SET, C.CS_PROP_READONLY, unsafe.Pointer(&ro), C.CS_UNUSED, nil)
	if rc != C.CS_SUCCEED {
		return nil, makeError(rc, "Failed to set readonly")
	}

	return nil
}
