package purego

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
)

var (
	_ driver.Connector = (*Connector)(nil)
)

type Connector struct {
	DSN            *libdsn.DsnInfo
	EnvChangeHooks []tds.EnvChangeHook
}

func NewConnector(dsn libdsn.DsnInfo) (driver.Connector, error) {
	return NewConnectorWithHooks(dsn)
}

func NewConnectorWithHooks(dsn libdsn.DsnInfo, hooks ...tds.EnvChangeHook) (driver.Connector, error) {
	connector := &Connector{
		DSN:            &dsn,
		EnvChangeHooks: hooks,
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
	return NewConnWithHooks(ctx, c.DSN, c.EnvChangeHooks)
}
