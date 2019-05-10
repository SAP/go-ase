package libtest

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-ase/libase/dsn"
)

var (
	testDsn                   *dsn.DsnInfo
	testDsnConnector          driver.Connector
	testDsnUserstore          *dsn.DsnInfo
	testDsnUserstoreConnector driver.Connector
)

// GetDBs opens a series of connections using different methods and
// returns them in a map.
func GetDBs() (map[string]*sql.DB, error) {
	direct, err := sql.Open("ase", testDsn.AsSimple())
	if err != nil {
		return nil, fmt.Errorf("Failed to open DB with %v: %v", testDsn, err)
	}

	userstoreDirect, err := sql.Open("ase", testDsnUserstore.AsSimple())
	if err != nil {
		direct.Close()
		return nil, fmt.Errorf("Failed to open DB with %v: %v", testDsnUserstore, err)
	}

	return map[string]*sql.DB{
		"direct":                 direct,
		"connector":              sql.OpenDB(testDsnConnector),
		"userstorekey_direct":    userstoreDirect,
		"userstorekey_connector": sql.OpenDB(testDsnUserstoreConnector),
	}, nil
}

// ConnectorCreator is the interface for function expected by InitDBs to
// initialize driver.Connectors.
type ConnectorCreator func(dsn.DsnInfo) (driver.Connector, error)

// InitDBs collects DSN information from the environment, creates a test
// database for the connection type and returns a function to tear down
// all created databases.
func InitDBs(connectorFn ConnectorCreator) (func(), error) {
	fnChan := make(chan func(), 4)
	defer close(fnChan)
	deferFn := func() {
		for fn := range fnChan {
			if fn != nil {
				fn()
			}
		}
	}

	var err error
	var fn func()

	// direct
	testDsn, fn, err = DSN(false)
	if err != nil {
		return deferFn, err
	}
	fnChan <- fn

	// connector
	testDsnConnectorDsn, fn, err := DSN(false)
	if err != nil {
		return deferFn, err
	}
	fnChan <- fn

	testDsnConnector, err = connectorFn(*testDsnConnectorDsn)
	if err != nil {
		return deferFn, fmt.Errorf("Failed to open connector: %v", err)
	}

	// userstorekey
	testDsnUserstore, fn, err = DSN(true)
	if err != nil {
		return deferFn, err
	}
	fnChan <- fn

	// userstorekey connector
	testDsnUserstoreConnectorDsn, fn, err := DSN(true)
	if err != nil {
		return deferFn, err
	}
	fnChan <- fn

	testDsnUserstoreConnector, err = connectorFn(*testDsnUserstoreConnectorDsn)
	if err != nil {
		return deferFn, fmt.Errorf("Failed to open connector: %v", err)
	}

	return deferFn, nil
}
