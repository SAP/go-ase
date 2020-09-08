// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

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

	err = libtest.RegisterDSN("username password", simpleDSN, ase.NewConnector)
	if err != nil {
		return fmt.Errorf("error setting up simple database: %w", err)
	}

	rc := m.Run()
	if rc != 0 {
		return fmt.Errorf("tests failed with %d", rc)
	}

	return nil
}

// Exact numeric integer
func TestInt8(t *testing.T)  { libtest.DoTestBigInt(t) }
func TestInt4(t *testing.T)  { libtest.DoTestInt(t) }
func TestInt2(t *testing.T)  { libtest.DoTestSmallInt(t) }
func TestInt1(t *testing.T)  { libtest.DoTestTinyInt(t) }
func TestUint8(t *testing.T) { libtest.DoTestUnsignedBigInt(t) }
func TestUint4(t *testing.T) { libtest.DoTestUnsignedInt(t) }
func TestUint2(t *testing.T) { libtest.DoTestUnsignedSmallInt(t) }

// Approximate numeric
func TestFlt8(t *testing.T) { libtest.DoTestFloat(t) }
func TestFlt4(t *testing.T) { libtest.DoTestReal(t) }

// Bit
func TestBit(t *testing.T) { libtest.DoTestBit(t) }

// Routines
func TestSQLTx(t *testing.T)   { libtest.DoTestSQLTx(t) }
func TestSQLExec(t *testing.T) { libtest.DoTestSQLExec(t) }
