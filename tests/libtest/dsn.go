// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package libtest

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"

	"github.com/SAP/go-ase/libase/libdsn"
)

// fromEnv reads an environment variable and returns the value.
//
// If the variable is not set in the environment a message is printed to
// stderr and os.Exit is called.
func fromEnv(name string) (string, error) {
	target, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("missing environment variable: %s", name)
	}

	return target, nil
}

// doCallbacks return true if CGO_CALLBACKS is set to 'yes', signaling
// that client-library callbacks should be used.
func doCallbacks() bool {
	val, ok := os.LookupEnv("CGO_CALLBACKS")
	if !ok {
		return false
	}

	return val == "yes"
}

// DSNFromEnv initializes a libdsn.Info and fills it with information
// from the environment.
func DSNFromEnv() (*libdsn.Info, error) {
	dsnInfo := libdsn.NewInfo()

	var err error
	dsnInfo.Host, err = fromEnv("ASE_HOST")
	if err != nil {
		return nil, err
	}

	dsnInfo.Port, err = fromEnv("ASE_PORT")
	if err != nil {
		return nil, err
	}

	dsnInfo.Username, err = fromEnv("ASE_USER")
	if err != nil {
		return nil, err
	}

	dsnInfo.Password, err = fromEnv("ASE_PASS")
	if err != nil {
		return nil, err
	}

	if doCallbacks() {
		dsnInfo.ConnectProps["cgo-callback-client"] = []string{"yes"}
		dsnInfo.ConnectProps["cgo-callback-server"] = []string{"yes"}
	}

	return dsnInfo, nil
}

// DSNUserstoreFromEnv initializes a dsn.Info and retrieves the
// userstorekey from the environment.
func DSNUserstoreFromEnv() (*libdsn.Info, error) {
	dsnInfo := libdsn.NewInfo()

	var err error

	dsnInfo.Userstorekey, err = fromEnv("ASE_USERSTOREKEY")
	if err != nil {
		return nil, err
	}

	if doCallbacks() {
		dsnInfo.ConnectProps["cgo-callback-client"] = []string{"yes"}
		dsnInfo.ConnectProps["cgo-callback-server"] = []string{"yes"}
	}

	return dsnInfo, nil
}

// DSN creates a new dsn.Info, sets up a new database and returns the
// Info and a function to tear down the database.
func DSN(userstore bool) (*libdsn.Info, func(), error) {
	var dsn *libdsn.Info
	var err error
	if !userstore {
		dsn, err = DSNFromEnv()
	} else {
		dsn, err = DSNUserstoreFromEnv()
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DSN from environment: %v", err)
	}

	err = SetupDB(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup database: %v", err)
	}

	fn := func() {
		err := TeardownDB(dsn)
		if err != nil {
			log.Printf("failed to drop database: %s", dsn.Database)
		}
	}

	return dsn, fn, nil
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
