// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

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
		log.Fatalf("%v", err)
		os.Exit(1)
	}
}

func testMain(m *testing.M) error {
	simpleDSN, simpleTeardown, err := libtest.DSN(false)
	if err != nil {
		return fmt.Errorf("error setting up simple DSN: %w", err)
	}
	defer simpleTeardown()

	err = libtest.RegisterDSN("username password", simpleDSN, cgo.NewConnector)
	if err != nil {
		return fmt.Errorf("error setting up simple database: %w", err)
	}

	userstoreDSN, userstoreTeardown, err := libtest.DSN(true)
	if err != nil {
		return fmt.Errorf("error setting up userstore DSN: %v", err)
	}
	defer userstoreTeardown()

	err = libtest.RegisterDSN("userstorekey", userstoreDSN, cgo.NewConnector)
	if err != nil {
		return fmt.Errorf("error setting up userstore database: %w", err)
	}

	cgo.GlobalServerMessageBroker.RegisterHandler(genMessageHandler())
	cgo.GlobalClientMessageBroker.RegisterHandler(genMessageHandler())

	rc := m.Run()
	if rc != 0 {
		return fmt.Errorf("Tests failed with %d", rc)
	}

	return nil
}

// Exact numeric integer
func TestBigInt(t *testing.T)           { libtest.DoTestBigInt(t) }
func TestInt(t *testing.T)              { libtest.DoTestInt(t) }
func TestSmallInt(t *testing.T)         { libtest.DoTestSmallInt(t) }
func TestTinyInt(t *testing.T)          { libtest.DoTestTinyInt(t) }
func TestUnsignedBigInt(t *testing.T)   { libtest.DoTestUnsignedBigInt(t) }
func TestUnsignedInt(t *testing.T)      { libtest.DoTestUnsignedInt(t) }
func TestUnsignedSmallInt(t *testing.T) { libtest.DoTestUnsignedSmallInt(t) }

// Exact numeric decimal
func TestDecimal(t *testing.T)     { libtest.DoTestDecimal(t) }
func TestDecimal10(t *testing.T)   { libtest.DoTestDecimal10(t) }
func TestDecimal380(t *testing.T)  { libtest.DoTestDecimal380(t) }
func TestDecimal3838(t *testing.T) { libtest.DoTestDecimal3838(t) }

// Approximate numeric
func TestFloat(t *testing.T) { libtest.DoTestFloat(t) }
func TestReal(t *testing.T)  { libtest.DoTestReal(t) }

// Money
func TestMoney(t *testing.T)  { libtest.DoTestMoney(t) }
func TestMoney4(t *testing.T) { libtest.DoTestMoney4(t) }

// Date and time
func TestDate(t *testing.T)          { libtest.DoTestDate(t) }
func TestTime(t *testing.T)          { libtest.DoTestTime(t) }
func TestSmallDateTime(t *testing.T) { libtest.DoTestSmallDateTime(t) }
func TestDateTime(t *testing.T)      { libtest.DoTestDateTime(t) }
func TestBigDateTime(t *testing.T)   { libtest.DoTestBigDateTime(t) }
func TestBigTime(t *testing.T)       { libtest.DoTestBigTime(t) }

// Character
func TestVarChar(t *testing.T)  { libtest.DoTestVarChar(t) }
func TestChar(t *testing.T)     { libtest.DoTestChar(t) }
func TestNChar(t *testing.T)    { libtest.DoTestNChar(t) }
func TestNVarChar(t *testing.T) { libtest.DoTestNVarChar(t) }
func TestText(t *testing.T)     { libtest.DoTestText(t) }
func TestUniChar(t *testing.T)  { libtest.DoTestUniChar(t) }
func TestUniText(t *testing.T)  { libtest.DoTestUniText(t) }

// Binary
func TestBinary(t *testing.T)    { libtest.DoTestBinary(t) }
func TestVarBinary(t *testing.T) { libtest.DoTestVarBinary(t) }

// Bit
func TestBit(t *testing.T) { libtest.DoTestBit(t) }

// Image
func TestImage(t *testing.T) { libtest.DoTestImage(t) }

// Routines
func TestSQLTx(t *testing.T)   { libtest.DoTestSQLTx(t) }
func TestSQLExec(t *testing.T) { libtest.DoTestSQLExec(t) }
