// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package purego

import (
	"fmt"
	"log"
	"testing"

	"github.com/SAP/go-ase/tests/libtest"
)

func TestMain(m *testing.M) {
	if err := testMain(m); err != nil {
		log.Fatal(err)
	}
}

func testMain(m *testing.M) error {
	simpleDSN, simpleTeardown, err := libtest.DSN(false)
	if err != nil {
		return fmt.Errorf("error setting up simple DSN: %w", err)
	}
	defer simpleTeardown()

	if err := libtest.RegisterDSN("username password", simpleDSN, NewConnector); err != nil {
		return fmt.Errorf("error setting up simple database: %w", err)
	}

	if rc := m.Run(); rc != 0 {
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

// Exact numeric decimal
func TestDecimal(t *testing.T)     { libtest.DoTestDecimal(t) }
func TestDecimal10(t *testing.T)   { libtest.DoTestDecimal10(t) }
func TestDecimal380(t *testing.T)  { libtest.DoTestDecimal380(t) }
func TestDecimal3838(t *testing.T) { libtest.DoTestDecimal3838(t) }

// Approximate numeric
func TestFlt8(t *testing.T) { libtest.DoTestFloat(t) }
func TestFlt4(t *testing.T) { libtest.DoTestReal(t) }

// Money
func TestMoney(t *testing.T)      { libtest.DoTestMoney(t) }
func TestShortmoney(t *testing.T) { libtest.DoTestMoney4(t) }

// Date and time
func TestDateN(t *testing.T)         { libtest.DoTestDate(t) }
func TestTimeN(t *testing.T)         { libtest.DoTestTime(t) }
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

// TODO func TestUniChar(t *testing.T)  { libtest.DoTestUniChar(t) }
// TODO func TestUniText(t *testing.T)  { libtest.DoTestUniText(t) }

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
