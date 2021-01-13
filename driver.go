// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/tds"
)

// Interface satisfaction checks.
var (
	_   driver.Driver        = (*Driver)(nil)
	_   driver.DriverContext = (*Driver)(nil)
	drv                      = &Driver{}
)

// DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

func init() {
	sql.Register(DriverName, drv)
}

// Driver implements the driver.Driver interface.
type Driver struct {
	envChangeHooks []tds.EnvChangeHook
	eedHooks       []tds.EEDHook
}

// Open implements the driver.Driver interface.
func (d Driver) Open(name string) (driver.Conn, error) {
	connector, err := d.OpenConnector(name)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error opening connector: %w", err)
	}

	return connector.Connect(context.Background())
}

// OpenConnector implements the driver.DriverContext interface.
func (d Driver) OpenConnector(name string) (driver.Connector, error) {
	dsnInfo, err := dsn.ParseDSN(name)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error parsing DSN: %w", err)
	}

	return NewConnector(dsnInfo)
}

// AddEnvChangeHooks gathers the envChangeHooks.
func AddEnvChangeHooks(fns ...tds.EnvChangeHook) error {
	for _, fn := range fns {
		if fn == nil {
			return fmt.Errorf("go-ase: Received nil EnvChangeHook: %#v", fns)
		}
	}

	drv.envChangeHooks = append(drv.envChangeHooks, fns...)
	return nil
}

// AddEEDHooks gathers the eedHooks.
func AddEEDHooks(fns ...tds.EEDHook) error {
	for _, fn := range fns {
		if fn == nil {
			return fmt.Errorf("go-ase: Received nil EEDHook: %#v", fns)
		}
	}

	drv.eedHooks = append(drv.eedHooks, fns...)
	return nil
}
