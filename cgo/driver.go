package cgo

/*
#cgo CFLAGS: -I${SRCDIR}/includes
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
	"time"

	"github.com/SAP/go-ase/libase"
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
	dsnInfo, err := libase.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN: %v", err)
	}

	return newConnection(*dsnInfo)
}

// needed to handle nil time values
type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nullTime *NullTime) Scan(value interface{}) error {
	nullTime.Time, nullTime.Valid = value.(time.Time)
	return nil
}

func (nullTime NullTime) Value() (driver.Value, error) {
	if !nullTime.Valid {
		return nil, nil
	}
	return nullTime.Time, nil
}

// needed to handle nil binary values
type NullBytes struct {
	Bytes []byte
	Valid bool
}

func (nullBytes *NullBytes) Scan(value interface{}) error {
	nullBytes.Bytes, nullBytes.Valid = value.([]byte)
	return nil
}

func (nullBytes NullBytes) Value() (driver.Value, error) {
	if !nullBytes.Valid {
		return nil, nil
	}
	return nullBytes.Bytes, nil
}
