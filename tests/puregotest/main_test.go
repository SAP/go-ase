package puregotest

import (
	"fmt"
	"log"
	"testing"

	ase "github.com/SAP/go-ase/purego"
	"github.com/SAP/go-ase/tests/libtest"
)

func TestMain(m *testing.M) {
	err := testMain(m)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func testMain(m *testing.M) error {
	simpleDSN, simpleTeardown, err := libtest.DSN(false)
	if err != nil {
		return fmt.Errorf("error setting up simple DSN: %w", err)
	}
	defer simpleTeardown()

	err = libtest.RegisterDSN("username password", *simpleDSN, ase.NewConnector)
	if err != nil {
		return fmt.Errorf("error setting up simple database: %w", err)
	}

	rc := m.Run()
	if rc != 0 {
		return fmt.Errorf("tests failed with %d", rc)
	}

	return nil
}
