package cgo

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/SAP/go-ase/libase"
)

// fromEnv reads an environment variable and returns the value.
//
// If the variable is not set in the environment a message is printed to
// stderr and os.Exit is called.
func fromEnv(name string) string {
	target, ok := os.LookupEnv(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "Missing environment variable: %s\n", name)
		os.Exit(1)
	}

	return target
}

// testDsn is a global DSN providing information for a running and
// available ASE to run tests against.
var testDsn = &libase.DsnInfo{}

// TestMain fills testDsn with information provided from the environment
// and triggers the tests.
func TestMain(m *testing.M) {
	testDsn.Host = fromEnv("ASE_HOST")
	testDsn.Port = fromEnv("ASE_PORT")
	testDsn.Username = fromEnv("ASE_USER")
	testDsn.Password = fromEnv("ASE_PASS")

	// Create new database to run tests in
	conn, err := newConnection(*testDsn)
	if err != nil {
		log.Printf("Failed to open connection to database: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	testDsn.Database = "db" + strconv.Itoa(rand.Int())

	_, err = conn.Exec("use master", nil)
	if err != nil {
		log.Printf("Failed to switch database context to master: %v", err)
		conn.Close()
		os.Exit(1)
	}

	_, err = conn.Exec("create database "+testDsn.Database, nil)
	conn.Close()
	if err != nil {
		log.Printf("Failed to create database %s: %v", testDsn.Database, err)
		os.Exit(1)
	}

	// Run test suite
	rc := m.Run()

	// Delete test database
	conn, err = newConnection(*testDsn)
	if err != nil {
		log.Printf("Failed to open connection to database: %v", err)
		os.Exit(1)
	}

	_, err = conn.Exec("use master", nil)
	if err != nil {
		log.Printf("Failed to switch database context to master: %v", err)
		conn.Close()
		os.Exit(1)
	}

	_, err = conn.Exec("drop database "+testDsn.Database, nil)
	conn.Close()
	if err != nil {
		log.Printf("Failed to drop database %s: %v", testDsn.Database, err)
		os.Exit(1)
	}

	os.Exit(rc)
}
