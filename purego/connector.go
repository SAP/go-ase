package purego

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
)

var (
	_ driver.Connector = (*Connector)(nil)
)

type Connector struct {
	DSN *libdsn.DsnInfo
}

func NewConnector(dsn *libdsn.DsnInfo) (*Connector, error) {
	connector := &Connector{
		DSN: dsn,
	}

	conn, err := connector.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error opening test connection: %w", err)
	}
	defer conn.Close()

	pinger, ok := conn.(driver.Pinger)
	if !ok {
		return nil, fmt.Errorf("received conn does not satisfy the pinger interface: %v", conn)
	}

	err = pinger.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error pinging database server: %w", err)
	}

	return connector, nil
}

func (c Connector) Driver() driver.Driver {
	return drv
}

func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	return NewConn(ctx, c.DSN)
}
