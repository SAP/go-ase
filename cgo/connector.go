package cgo

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
)

var _ driver.Connector = (*connector)(nil)

type connector struct {
	driverCtx *csContext
	dsn       libdsn.DsnInfo
}

// NewConnector returns a driver.Connector which can be passed to
// sql.OpenDB.
func NewConnector(dsn libdsn.DsnInfo) (driver.Connector, error) {
	driverCtx, err := newCsContext(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize context: %v", err)
	}

	c := &connector{
		driverCtx: driverCtx,
		dsn:       dsn,
	}

	conn, err := c.Connect(context.Background())
	if err != nil {
		driverCtx.drop()
		return nil, fmt.Errorf("Failed to open connection: %v", err)
	}

	defer func() {
		// In- and decrease connections count before and after closing
		// connection to prevent the context being deallocated.
		driverCtx.connections++
		conn.Close()
		driverCtx.connections--
	}()

	return c, nil
}

func (connector *connector) Connect(ctx context.Context) (driver.Conn, error) {
	connChan := make(chan driver.Conn, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := newConnection(connector.driverCtx, connector.dsn)
		connChan <- conn
		close(connChan)
		errChan <- err
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		defer func() {
			conn := <-connChan
			if conn != nil {
				conn.Close()
			}
		}()
		return nil, ctx.Err()
	case err := <-errChan:
		conn := <-connChan
		return conn, err
	}
}

func (connector connector) Driver() driver.Driver {
	return drv
}
