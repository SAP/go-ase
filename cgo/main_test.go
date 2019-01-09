package cgo

import (
	"fmt"
	"os"
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

	os.Exit(m.Run())
}
