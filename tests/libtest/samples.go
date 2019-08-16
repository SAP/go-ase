package libtest

import (
	"math"
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
