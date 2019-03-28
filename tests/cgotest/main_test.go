package cgotest

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/tests/libtest"
)

func TestMain(m *testing.M) {
	err := testMain(m)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}

func testMain(m *testing.M) error {
	fn, err := libtest.InitDBs(cgo.NewConnector)
	if err != nil {
		return fmt.Errorf("Failed to initialize databases: %v", err)
	}
	defer fn()

	rc := m.Run()
	if rc != 0 {
		return fmt.Errorf("Tests failed with %d", rc)
	}

	return nil
}

func TestIn64(t *testing.T) {
	libtest.DoTestInt64(t)
}
