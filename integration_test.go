// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration
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

// Nullable exact numeric integer
func TestNullInt8(t *testing.T)  { integration.DoTestNullBigInt(t) }
func TestNullInt4(t *testing.T)  { integration.DoTestNullInt(t) }
func TestNullInt2(t *testing.T)  { integration.DoTestNullSmallInt(t) }
func TestNullInt1(t *testing.T)  { integration.DoTestNullTinyInt(t) }
func TestNullUint8(t *testing.T) { integration.DoTestNullUnsignedBigInt(t) }
func TestNullUint4(t *testing.T) { integration.DoTestNullUnsignedInt(t) }
func TestNullUint2(t *testing.T) { integration.DoTestNullUnsignedSmallInt(t) }

// Exact numeric decimal
func TestDecimal(t *testing.T)     { integration.DoTestDecimal(t) }
func TestDecimal10(t *testing.T)   { integration.DoTestDecimal10(t) }
func TestDecimal380(t *testing.T)  { integration.DoTestDecimal380(t) }
func TestDecimal3838(t *testing.T) { integration.DoTestDecimal3838(t) }

// Nullable exact numeric decimal
func TestNullDecimal(t *testing.T) { integration.DoTestNullDecimal(t) }

// Approximate numeric
func TestFlt8(t *testing.T) { integration.DoTestFloat(t) }
func TestFlt4(t *testing.T) { integration.DoTestReal(t) }

// Nullable approximate numeric
func TestNullFlt8(t *testing.T) { integration.DoTestNullFloat(t) }
func TestNullFlt4(t *testing.T) { integration.DoTestNullReal(t) }

// Money
func TestMoney(t *testing.T)      { integration.DoTestMoney(t) }
func TestShortmoney(t *testing.T) { integration.DoTestMoney4(t) }

// Nullable Money
func TestNullMoney(t *testing.T)      { integration.DoTestNullMoney(t) }
func TestNullShortmoney(t *testing.T) { integration.DoTestNullMoney4(t) }

// Date and time
func TestDateN(t *testing.T)         { integration.DoTestDate(t) }
func TestTimeN(t *testing.T)         { integration.DoTestTime(t) }
func TestBigTime(t *testing.T)       { integration.DoTestBigTime(t) }
func TestSmallDateTime(t *testing.T) { integration.DoTestSmallDateTime(t) }
func TestDateTime(t *testing.T)      { integration.DoTestDateTime(t) }
func TestBigDateTime(t *testing.T)   { integration.DoTestBigDateTime(t) }

// Nullable Date and time
func TestNullDate(t *testing.T)          { integration.DoTestNullDate(t) }
func TestNullTime(t *testing.T)          { integration.DoTestNullTime(t) }
func TestNullBigTime(t *testing.T)       { integration.DoTestNullBigTime(t) }
func TestNullSmallDateTime(t *testing.T) { integration.DoTestNullSmallDateTime(t) }
func TestNullDateTime(t *testing.T)      { integration.DoTestNullDateTime(t) }
func TestNullBigDateTime(t *testing.T)   { integration.DoTestNullBigDateTime(t) }

// Character
func TestChar(t *testing.T)     { integration.DoTestChar(t) }
func TestNChar(t *testing.T)    { integration.DoTestNChar(t) }
func TestVarChar(t *testing.T)  { integration.DoTestVarChar(t) }
func TestNVarChar(t *testing.T) { integration.DoTestNVarChar(t) }

// Nullable Character
func TestNullChar(t *testing.T)     { integration.DoTestNullChar(t) }
func TestNullNChar(t *testing.T)    { integration.DoTestNullNChar(t) }
func TestNullVarChar(t *testing.T)  { integration.DoTestNullVarChar(t) }
func TestNullNVarChar(t *testing.T) { integration.DoTestNullNVarChar(t) }

// TODO func TestText(t *testing.T)     { integration.DoTestText(t) }
// TODO func TestUniChar(t *testing.T)  { integration.DoTestUniChar(t) }
// TODO func TestUniText(t *testing.T)  { integration.DoTestUniText(t) }

// Binary
func TestBinary(t *testing.T)    { integration.DoTestBinary(t) }
func TestVarBinary(t *testing.T) { integration.DoTestVarBinary(t) }

// Nullable Binary
func TestNullBinary(t *testing.T)    { integration.DoTestNullBinary(t) }
func TestNullVarBinary(t *testing.T) { integration.DoTestNullVarBinary(t) }

// Bit
func TestBit(t *testing.T) { integration.DoTestBit(t) }

// Image
func TestImage(t *testing.T) { integration.DoTestImage(t) }

// Routines
func TestSQLTx(t *testing.T)   { integration.DoTestSQLTx(t) }
func TestSQLExec(t *testing.T) { integration.DoTestSQLExec(t) }
