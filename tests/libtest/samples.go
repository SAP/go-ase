package libtest

import (
	"math"

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

//go:generate go run ./gen_type.go Decimal10 github.com/SAP/go-ase/libase/*types.Decimal -columndef decimal(1,0) -convert github.com/SAP/go-ase/libase/types.NewDecimalString -compare compareDecimal
var samplesDecimal10 = []string{"0", "1", "9"}

//go:generate go run ./gen_type.go Decimal380 github.com/SAP/go-ase/libase/*types.Decimal -columndef decimal(38,0) -convert github.com/SAP/go-ase/libase/types.NewDecimalString -compare compareDecimal
var samplesDecimal380 = []string{"99999999999999999999999999999999999999"}

//go:generate go run ./gen_type.go Decimal3838 github.com/SAP/go-ase/libase/*types.Decimal -columndef decimal(38,38) -convert github.com/SAP/go-ase/libase/types.NewDecimalString -compare compareDecimal
var samplesDecimal3838 = []string{".99999999999999999999999999999999999999"}

//go:generate go run ./gen_type.go Decimal github.com/SAP/go-ase/libase/*types.Decimal -columndef "decimal(38,19)" -convert github.com/SAP/go-ase/libase/types.NewDecimalString -compare compareDecimal
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
	if recv.String() != expect.String() {
		return true
	}
	return false
}

//go:generate go run ./gen_type.go Float float64
// TODO: -null database/sql.NullFloat64
var samplesFloat = []float64{math.SmallestNonzeroFloat64, math.MaxFloat64}

//go:generate go run ./gen_type.go Real float64
// TODO: -null database/sql.NullFloat64
var samplesReal = []float64{
	math.SmallestNonzeroFloat32,
	math.MaxFloat32,
}
