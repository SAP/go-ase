package libtests

import (
	"fmt"
	"log"
	"os"

	"github.com/SAP/go-ase/libase/dsn"
)

// fromEnv reads an environment variable and returns the value.
//
// If the variable is not set in the environment a message is printed to
// stderr and os.Exit is called.
func fromEnv(name string) (string, error) {
	target, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("Missing environment variable: %s", name)
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

	if val == "yes" {
		return true
	}
	return false
}

func DSNFromEnv() (*dsn.DsnInfo, error) {
	dsnInfo := dsn.NewDsnInfo()

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

func DSNUserstoreFromEnv() (*dsn.DsnInfo, error) {
	dsnInfo := dsn.NewDsnInfo()

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

func DSN(userstore bool) (*dsn.DsnInfo, func(), error) {
	var dsn *dsn.DsnInfo
	var err error
	if !userstore {
		dsn, err = DSNFromEnv()
	} else {
		dsn, err = DSNUserstoreFromEnv()
	}

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create DSN from environment: %v", err)
	}

	err = SetupDB(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to setup database: %v", err)
	}

	fn := func() {
		log.Printf("Dropping db %s", dsn.Database)
		err := TeardownDB(dsn)
		if err != nil {
			log.Printf("Failed to drop database: %s", dsn.Database)
		}
	}

	return dsn, fn, nil
}
