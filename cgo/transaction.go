package cgo

//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"unsafe"

	"github.com/SAP/go-ase/libase"
)

// transaction is the struct which represents a database transaction.
type transaction struct {
	conn *connection
	// ASE does not support read-only transactions - the connection
	// itself however can be set as read-only.
	// readonlyPreTx signals if a connection was marked as read-only
	// before the transaction startet.
	readonlyPreTx C.CS_INT
	// readonlyNeedsReset is true if the read-only option passed to
	// BeginTx differs from the read-only property of the connection.
	readonlyNeedsReset bool
}

// Interface satisfaction checks
var _ driver.Tx = (*transaction)(nil)

func (conn *connection) Begin() (driver.Tx, error) {
	return conn.beginTx(driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelDefault), ReadOnly: false})
}

func (conn *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	recvTx := make(chan driver.Tx, 1)
	recvErr := make(chan error, 1)
	go func() {
		tx, err := conn.beginTx(opts)
		recvTx <- tx
		close(recvTx)
		recvErr <- err
		close(recvErr)
	}()

	for {
		select {
		case <-ctx.Done():
			// Context exits early, Tx will still be created and
			// initialized; read and rollback
			go func() {
				tx := <-recvTx
				if tx != nil {
					tx.Rollback()
				}
			}()
			return nil, ctx.Err()
		case tx := <-recvTx:
			if tx != nil {
				return tx, nil
			}
		case err := <-recvErr:
			if err != nil {
				return nil, err
			}
		}
	}
}

func (conn *connection) beginTx(opts driver.TxOptions) (driver.Tx, error) {
	isolationLevel, err := libase.IsolationLevelFromGo(sql.IsolationLevel(opts.Isolation))
	if err != nil {
		return nil, err
	}

	tx := &transaction{conn, C.CS_FALSE, false}

	_, err = tx.conn.Exec("BEGIN TRANSACTION", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to start transaction: %v", err)
	}

	_, err = tx.conn.Exec("SET TRANSACTION ISOLATION LEVEL ?", []driver.Value{int(isolationLevel)})
	if err != nil {
		return nil, fmt.Errorf("Failed to set isolation level for transaction: %v", err)
	}

	var currentReadOnly C.CS_INT = C.CS_FALSE
	retval := C.ct_con_props(tx.conn.conn, C.CS_GET, C.CS_PROP_READONLY, unsafe.Pointer(&currentReadOnly), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to retrieve readonly property")
	}

	var targetReadOnly C.CS_INT = C.CS_FALSE
	if opts.ReadOnly {
		targetReadOnly = C.CS_TRUE
	}

	if currentReadOnly != targetReadOnly {
		tx.readonlyPreTx = currentReadOnly
		tx.readonlyNeedsReset = true

		err = tx.setRO(targetReadOnly)
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
		err := tx.setRO(tx.readonlyPreTx)
		if err != nil {
			return err
		}
	}
	tx.conn = nil
	return nil
}

func (tx *transaction) setRO(ro C.CS_INT) error {
	retval := C.ct_con_props(tx.conn.conn, C.CS_SET, C.CS_PROP_READONLY, unsafe.Pointer(&ro), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Failed to set readonly")
	}

	return nil
}
