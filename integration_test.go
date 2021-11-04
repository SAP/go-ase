// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package ase

import (
	"context"
	"database/sql/driver"
	"fmt"
	"log"
	"testing"

	"github.com/SAP/go-dblib/integration"
)

func TestMain(m *testing.M) {
	if err := testMain(m); err != nil {
		log.Fatal(err)
	}
}

func testMain(m *testing.M) error {
	info, err := NewInfoWithEnv()
	if err != nil {
		return err
	}

	if err := integration.SetupDB(context.Background(), info, "test"+integration.RandomNumber()); err != nil {
		return err
	}

	defer func() {
		if err := integration.TeardownDB(context.Background(), info); err != nil {
			log.Printf("error dropping database %q: %v", info.Database, err)
		}
	}()

	newConnectorFn := func(info interface{}) (driver.Connector, error) {
		return NewConnector(info.(*Info))
	}

	if err := integration.RegisterDSN("username password", info, newConnectorFn); err != nil {
		return fmt.Errorf("error setting up simple database: %w", err)
	}

	if rc := m.Run(); rc != 0 {
		return fmt.Errorf("tests failed with %d", rc)
	}

	return nil
}

// Exact numeric integer
func TestInt8(t *testing.T)  { integration.DoTestBigInt(t) }
func TestInt4(t *testing.T)  { integration.DoTestInt(t) }
func TestInt2(t *testing.T)  { integration.DoTestSmallInt(t) }
func TestInt1(t *testing.T)  { integration.DoTestTinyInt(t) }
func TestUint8(t *testing.T) { integration.DoTestUnsignedBigInt(t) }
func TestUint4(t *testing.T) { integration.DoTestUnsignedInt(t) }
func TestUint2(t *testing.T) { integration.DoTestUnsignedSmallInt(t) }

// Exact numeric decimal
func TestDecimal(t *testing.T)     { integration.DoTestDecimal(t) }
func TestDecimal10(t *testing.T)   { integration.DoTestDecimal10(t) }
func TestDecimal380(t *testing.T)  { integration.DoTestDecimal380(t) }
func TestDecimal3838(t *testing.T) { integration.DoTestDecimal3838(t) }

// Approximate numeric
func TestFlt8(t *testing.T) { integration.DoTestFloat(t) }
func TestFlt4(t *testing.T) { integration.DoTestReal(t) }

// Money
func TestMoney(t *testing.T)      { integration.DoTestMoney(t) }
func TestShortmoney(t *testing.T) { integration.DoTestMoney4(t) }

// Date and time
func TestDateN(t *testing.T)         { integration.DoTestDate(t) }
func TestTimeN(t *testing.T)         { integration.DoTestTime(t) }
func TestSmallDateTime(t *testing.T) { integration.DoTestSmallDateTime(t) }
func TestDateTime(t *testing.T)      { integration.DoTestDateTime(t) }
func TestBigDateTime(t *testing.T)   { integration.DoTestBigDateTime(t) }
func TestBigTime(t *testing.T)       { integration.DoTestBigTime(t) }

// Character
func TestVarChar(t *testing.T)  { integration.DoTestVarChar(t) }
func TestChar(t *testing.T)     { integration.DoTestChar(t) }
func TestNChar(t *testing.T)    { integration.DoTestNChar(t) }
func TestNVarChar(t *testing.T) { integration.DoTestNVarChar(t) }

// TODO func TestText(t *testing.T)     { integration.DoTestText(t) }
// TODO func TestUniChar(t *testing.T)  { integration.DoTestUniChar(t) }
// TODO func TestUniText(t *testing.T)  { integration.DoTestUniText(t) }

// Binary
func TestBinary(t *testing.T)    { integration.DoTestBinary(t) }
func TestVarBinary(t *testing.T) { integration.DoTestVarBinary(t) }

// Bit
func TestBit(t *testing.T) { integration.DoTestBit(t) }

// Image
func TestImage(t *testing.T) { integration.DoTestImage(t) }

// Routines
func TestSQLTx(t *testing.T)   { integration.DoTestSQLTx(t) }
func TestSQLExec(t *testing.T) { integration.DoTestSQLExec(t) }
