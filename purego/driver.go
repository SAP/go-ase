package purego

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
)

var (
	_   driver.Driver        = (*Driver)(nil)
	_   driver.DriverContext = (*Driver)(nil)
	drv                      = &Driver{}
)

const DriverName = "ase"

func init() {
	sql.Register(DriverName, drv)
}

type Driver struct {
}

func (d Driver) Open(name string) (driver.Conn, error) {
	connector, err := d.OpenConnector(name)
	if err != nil {
		return nil, fmt.Errorf("error opening connector: %w", err)
	}

	return connector.Connect(context.Background())
}

func (d Driver) OpenConnector(name string) (driver.Connector, error) {
	dsnInfo, err := libdsn.ParseDSN(name)
	if err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	return NewConnector(*dsnInfo)
}
