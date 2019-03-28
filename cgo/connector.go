package cgo

import (
	"context"
	"database/sql/driver"
	"fmt"

	libdsn "github.com/SAP/go-ase/libase/dsn"
)

var _ driver.Connector = (*connector)(nil)

type connector struct {
	driverCtx *csContext
	dsn       libdsn.DsnInfo
}

func NewConnector(dsn libdsn.DsnInfo) (driver.Connector, error) {
	driverCtx, err := newCsContext(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize context: %v")
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
		driverCtx.connections += 1
		conn.Close()
		driverCtx.connections -= 1
	}()

	return c, nil
}

func (connector *connector) Connect(ctx context.Context) (driver.Conn, error) {
	connChan := make(chan driver.Conn, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := newConnection(connector.driverCtx, connector.dsn)
		if err != nil {
			errChan <- err
		} else {
			connChan <- conn
		}

		close(connChan)
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
	case conn := <-connChan:
		return conn, nil
	case err := <-errChan:
		return nil, err
	}
}

func (connector connector) Driver() driver.Driver {
	return drv
}
