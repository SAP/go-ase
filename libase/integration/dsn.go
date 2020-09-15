// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"

	"github.com/SAP/go-ase/libase/libdsn"
)

// doCallbacks return true if CGO_CALLBACKS is set to 'yes', signaling
// that client-library callbacks should be used.
func doCallbacks() bool {
	val, ok := os.LookupEnv("CGO_CALLBACKS")
	if !ok {
		return false
	}

	return val == "yes"
}

// DSN creates a new dsn.Info, sets up a new database and returns the
// Info and a function to tear down the database.
func DSN(userstore bool) (*libdsn.Info, func(), error) {
	info, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return nil, nil, fmt.Errorf("error reading DSN info from env: %w", err)
	}

	if !userstore {
		info.Userstorekey = ""
	} else {
		info.Host = ""
		info.Port = ""
		info.Username = ""
		info.Password = ""
	}

	if doCallbacks() {
		info.ConnectProps["cgo-callback-client"] = []string{"yes"}
		info.ConnectProps["cgo-callback-server"] = []string{"yes"}
	}

	err = SetupDB(info)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup database: %w", err)
	}

	fn := func() {
		err := TeardownDB(info)
		if err != nil {
			log.Printf("failed to drop database: %s", info.Database)
		}
	}

	return info, fn, nil
}

// genSQLDBFn is the signature of functions stored in the genSQLDBMap.
type genSQLDBFn func() (*sql.DB, error)

// genSQLDBMap maps abstract names to functions, which are expected to
// return unique sql.DBs.
type genSQLDBMap map[string]genSQLDBFn

var sqlDBMap = make(genSQLDBMap)

// ConnectorCreator is the interface for function expected by InitDBs to
// initialize driver.Connectors.
type ConnectorCreator func(*libdsn.Info) (driver.Connector, error)

// RegisterDSN registers at least one new genSQLDBFn in genSQLDBMap
// based on sql.Open.
// If connectorFn is non-nil a second genSQLDBFn is stored with the
// suffix `connector`.
func RegisterDSN(name string, info *libdsn.Info, connectorFn ConnectorCreator) error {
	sqlDBMap[name] = func() (*sql.DB, error) {
		db, err := sql.Open("ase", info.AsSimple())
		if err != nil {
			return nil, err
		}
		return db, nil
	}

	if connectorFn != nil {
		sqlDBMap[name+" connector"] = func() (*sql.DB, error) {
			connector, err := connectorFn(info)
			if err != nil {
				return nil, err
			}

			return sql.OpenDB(connector), nil
		}
	}

	return nil
}
