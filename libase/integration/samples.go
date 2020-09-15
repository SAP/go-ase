// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"math"
	"strings"
	"time"

	"github.com/SAP/go-ase/libase/types"
)

//go:generate go run ./gen_type.go BigInt int64
// TODO: -null database/sql.NullInt64
var samplesBigInt = []int64{
	math.MinInt64, math.MaxInt64,
	-5000, -100, 0, 100, 5000,
}

//go:generate go run ./gen_type.go Int int32
var samplesInt = []int32{
	math.MinInt32, math.MaxInt32,
	-5000, -100, 0, 100, 5000,
}

//go:generate go run ./gen_type.go SmallInt int16
var samplesSmallInt = []int16{-32768, 0, 32767}

//go:generate go run ./gen_type.go TinyInt uint8
var samplesTinyInt = []uint8{0, 255}

//go:generate go run ./gen_type.go UnsignedBigInt uint64 -columndef "unsigned bigint"
var samplesUnsignedBigInt = []uint64{0, 1000, 5000, 150000, 123456789, math.MaxUint32 + 1}

//go:generate go run ./gen_type.go UnsignedInt uint32 -columndef "unsigned int"
var samplesUnsignedInt = []uint32{0, 1000, 5000, 150000, 123456789, math.MaxUint32}

//go:generate go run ./gen_type.go UnsignedSmallInt uint16 -columndef "unsigned smallint"
var samplesUnsignedSmallInt = []uint16{0, 65535}

func convertDecimal10(sample string) (*types.Decimal, error) {
	return types.NewDecimalString(1, 0, sample)
}

//go:generate go run ./gen_type.go Decimal10 github.com/SAP/go-ase/libase/*types.Decimal -columndef decimal(1,0) -convert convertDecimal10 -compare compareDecimal
var samplesDecimal10 = []string{"0", "1", "9"}

func convertDecimal380(sample string) (*types.Decimal, error) {
	return types.NewDecimalString(38, 0, sample)
}

//go:generate go run ./gen_type.go Decimal380 github.com/SAP/go-ase/libase/*types.Decimal -columndef decimal(38,0) -convert convertDecimal380 -compare compareDecimal
var samplesDecimal380 = []string{"99999999999999999999999999999999999999"}

func convertDecimal3838(sample string) (*types.Decimal, error) {
	return types.NewDecimalString(38, 38, sample)
}

//go:generate go run ./gen_type.go Decimal3838 github.com/SAP/go-ase/libase/*types.Decimal -columndef decimal(38,38) -convert convertDecimal3838 -compare compareDecimal
var samplesDecimal3838 = []string{".99999999999999999999999999999999999999"}

func convertDecimal3819(sample string) (*types.Decimal, error) {
	return types.NewDecimalString(38, 19, sample)
}

//go:generate go run ./gen_type.go Decimal github.com/SAP/go-ase/libase/*types.Decimal -columndef "decimal(38,19)" -convert convertDecimal3819 -compare compareDecimal
var samplesDecimal = []string{
	// ASE max
	"1234567890123456789",
	"9999999999999999999",
	"-1234567890123456789",
	"-9999999999999999999",
	// ASE min
	".1234567890123456789",
	".9999999999999999999",
	"-.1234567890123456789",
	"-.9999999999999999999",
	// default
	"0",
	// arbitrary
	"1234.5678",
}

func compareDecimal(recv, expect *types.Decimal) bool {
	return !expect.Cmp(*recv)
}

//go:generate go run ./gen_type.go Float float64
// TODO: -null database/sql.NullFloat64
var samplesFloat = []float64{
	-math.SmallestNonzeroFloat64,
	math.SmallestNonzeroFloat64,
	-1000,
	1000,
	-math.MaxFloat64,
	math.MaxFloat64,
}

//go:generate go run ./gen_type.go Real float32
// TODO: -null database/sql.NullFloat32
var samplesReal = []float32{
	-math.SmallestNonzeroFloat32,
	math.SmallestNonzeroFloat32,
	-1000,
	1000,
	-math.MaxFloat32,
	math.MaxFloat32,
}

func convertMoney(sample string) (*types.Decimal, error) {
	return types.NewDecimalString(types.ASEMoneyPrecision, types.ASEMoneyScale, sample)
}

//go:generate go run ./gen_type.go Money github.com/SAP/go-ase/libase/*types.Decimal -convert convertMoney -compare compareDecimal
var samplesMoney = []string{
	// ASE min
	"-922337203685477.5807",
	// ASE max
	"922337203685477.5807",
	// default
	"0.0",
	// arbitrary
	"1234.5678",
}

func convertSmallMoney(sample string) (*types.Decimal, error) {
	return types.NewDecimalString(types.ASEShortMoneyPrecision, types.ASEShortMoneyScale, sample)
}

//go:generate go run ./gen_type.go Money4 github.com/SAP/go-ase/libase/*types.Decimal -columndef smallmoney -convert convertSmallMoney -compare compareDecimal
var samplesMoney4 = []string{
	// ASE min
	"-214748.3648",
	// ASE max
	"214748.3647",
	// default
	"0.0",
	// arbitrary
	"1234.5678",
}

//go:generate go run ./gen_type.go Date time.Time
var samplesDate = []time.Time{
	// Sybase & Golang zero value
	time.Time{},
	// Sybase max
	time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC),
}

//go:generate go run ./gen_type.go Time time.Time
var samplesTime = []time.Time{
	// Sybase & Golang zero-value; 00:00:00.00
	time.Time{},
	// 13:15:55.123
	time.Date(1, time.January, 1, 13, 15, 55, 123000000, time.UTC),
	// Sybase max: 23:59:59.990
	time.Date(1, time.January, 1, 23, 59, 59, 996000000, time.UTC),
}

//go:generate go run ./gen_type.go SmallDateTime time.Time
var samplesSmallDateTime = []time.Time{
	// Sybase zero-value; January 1, 1900 Midnight
	time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
	// Sybase max: 06.06.2079 23:59
	time.Date(2079, time.June, 6, 23, 59, 0, 0, time.UTC),
}

//go:generate go run ./gen_type.go DateTime time.Time
var samplesDateTime = []time.Time{
	// Sybase min: January 1, 1753 Midnight
	time.Date(1753, time.January, 1, 0, 0, 0, 0, time.UTC),
	// Sybase zero-value; January 1, 1900 Midnight
	time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
	// Sybase max: 31.12.9999 23:59:59.996
	time.Date(9999, time.December, 31, 23, 59, 59, 996000000, time.UTC),
}

//go:generate go run ./gen_type.go BigDateTime time.Time
// TODO: -null github.com/SAP/go-ase/libase/types.NullTime
var samplesBigDateTime = []time.Time{
	// Sybase & Golang zero-value; January 1, 0001 Midnight
	time.Time{},
	time.Date(2019, time.March, 29, 9, 26, 0, 0, time.UTC),
	// Sybase max
	time.Date(9999, time.December, 31, 23, 59, 59, 999999000, time.UTC),
}

//go:generate go run ./gen_type.go BigTime time.Time
var samplesBigTime = []time.Time{
	// Sybase & Golang zero-value; 00:00:00.00
	time.Time{},
	// Sybase max: 23:59:59.999999
	time.Date(1, time.January, 1, 23, 59, 59, 999999000, time.UTC),
}

//go:generate go run ./gen_type.go VarChar string -columndef "varchar(13) null" -compare compareChar
// TODO: -null database/sql.NullString
var samplesVarChar = samplesChar

//go:generate go run ./gen_type.go Char string -columndef "char(13) null" -compare compareChar
// TODO: -null database/sql.NullString
var samplesChar = []string{"", "test", "a longer test"}

//go:generate go run ./gen_type.go NChar string -columndef "nchar(13) null"  -compare compareChar
// TODO: -null database/sql.NullString
var samplesNChar = samplesChar

//go:generate go run ./gen_type.go NVarChar string -columndef "nvarchar(13) null" -compare compareChar
// TODO: -null database/sql.NullString
var samplesNVarChar = samplesChar

func compareChar(recv, expect string) bool {
	return strings.Compare(strings.TrimSpace(recv), expect) != 0
}

//go:generate go run ./gen_type.go Binary []byte -columndef binary(13) -compare compareBinary
// TODO: -null github.com/SAP/go-ase/libase/types.NullBinary
var samplesBinary = [][]byte{
	[]byte("test"),
	[]byte("a longer test"),
}

//go:generate go run ./gen_type.go VarBinary []byte -columndef varbinary(13) -compare compareBinary
// TODO: -null github.com/SAP/go-ase/libase/types.NullBinary
var samplesVarBinary = samplesBinary

func compareBinary(recv, expect []byte) bool {
	return !bytes.Equal(bytes.Trim(recv, "\x00"), expect)
}

//go:generate go run ./gen_type.go Bit bool
// Cannot be nulled
var samplesBit = []bool{true, false}

//go:generate go run ./gen_type.go Image []byte -compare compareBinary
// TODO: -null github.com/SAP/go-ase/libase/types.NullBinary
var samplesImage = [][]byte{[]byte("test"), []byte("a longer test")}

// TODO: Separate null test, ctlib transforms empty value to null
//go:generate go run ./gen_type.go UniChar string -columndef "unichar(30) null" -compare compareChar
// TODO: -null database/sql.NullString
var samplesUniChar = []string{"", "not a unicode example"}

// TODO: Separate null test, ctlib transforms empty value to null
//go:generate go run ./gen_type.go Text string -columndef "text null" -compare compareChar
// TODO: -null database/sql.NullString
var samplesText = []string{"", "a long text"}

//go:generate go run ./gen_type.go UniText string -columndef unitext -compare compareChar
// TODO: -null database/sql.NullString
var samplesUniText = []string{"not a unicode example", "another not unicode example"}
