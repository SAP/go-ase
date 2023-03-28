// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-dblib/tds"
)

// Interface satisfaction checks
var (
	_ driver.Connector = (*Connector)(nil)
)

// Connector implements the driver.Connector interface.
type Connector struct {
	Info           *Info
	EnvChangeHooks []tds.EnvChangeHook
	EEDHooks       []tds.EEDHook
}

// NewConnector returns a new connector with the passed configuration.
func NewConnector(info *Info) (driver.Connector, error) {
	return NewConnectorWithHooks(info, nil, nil)
}

// NewConnectorWithHooks returns a new connector with the passed
// configuration.
func NewConnectorWithHooks(info *Info, envChangeHooks []tds.EnvChangeHook, eedHooks []tds.EEDHook) (driver.Connector, error) {
	connector := &Connector{
		Info: info,
	}

	conn, err := connector.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error opening test connection: %w", err)
	}

	if err := conn.Close(); err != nil {
		return nil, fmt.Errorf("error closing test connection: %w", err)
	}

	// Set the hooks after validating the connection otherwise hooks
	// would get called during the test connection.
	connector.EnvChangeHooks = envChangeHooks
	connector.EEDHooks = eedHooks

	return connector, nil
}

// Driver implements the driver.Connector interface.
func (c Connector) Driver() driver.Driver {
	return drv
}

// Connect implements the driver.Connector interface.
func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	return NewConnWithHooks(ctx, c.Info, c.EnvChangeHooks, c.EEDHooks)
}
