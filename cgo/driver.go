// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// Package cgo is the cgo implementation of go-ase.
package cgo

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -lsybct64 -lsybct_r64 -lsybcs_r64 -lsybtcl_r64 -lsybcomn_r64 -lsybintl_r64 -lsybunic64
#cgo LDFLAGS: -Wl,-rpath,\$ORIGIN/../lib
#include <stdlib.h>
#include "ctlib.h"
#include "bridge.h"
*/
import "C"
import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
)

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

// drv is the struct on which we later call Open() to get a connection.
type aseDrv struct{}

var (
	// Interface satisfaction checks
	_   driver.Driver = (*aseDrv)(nil)
	drv               = &aseDrv{}
)

func init() {
	sql.Register(DriverName, drv)
}

func (d *aseDrv) Open(dsn string) (driver.Conn, error) {
	dsnInfo, err := libdsn.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN: %w", err)
	}

	return NewConnection(nil, dsnInfo)
}

func (d *aseDrv) OpenConnector(dsn string) (driver.Connector, error) {
	dsnInfo, err := libdsn.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN: %w", err)
	}

	return NewConnector(dsnInfo)
}
