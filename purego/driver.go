// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
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
	envChangeHooks []tds.EnvChangeHook
}

func (d Driver) Open(name string) (driver.Conn, error) {
	connector, err := d.OpenConnector(name)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error opening connector: %w", err)
	}

	return connector.Connect(context.Background())
}

func (d Driver) OpenConnector(name string) (driver.Connector, error) {
	dsnInfo, err := libdsn.ParseDSN(name)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error parsing DSN: %w", err)
	}

	return NewConnector(dsnInfo)
}

func AddEnvChangeHooks(fns ...tds.EnvChangeHook) error {
	for _, fn := range fns {
		if fn == nil {
			return fmt.Errorf("go-ase: Received nil EnvChangeHook: %#v", fns)
		}
	}

	drv.envChangeHooks = append(drv.envChangeHooks, fns...)
	return nil
}
