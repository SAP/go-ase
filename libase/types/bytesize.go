package types

// ByteSizes is a map that maps a DataType to the length of their byte
// representation.
// Only fixed-length DataTypes are listed.
// Fixed-length DataTypes that are nullable are not listed.
var ByteSizes = map[DataType]int{
	BIT:        1,
	DATE:       4,
	DATETIME:   8,
	FLT4:       4,
	FLT8:       8,
	INT1:       1,
	INT2:       2,
	INT4:       4,
	INT8:       8,
	INTERVAL:   8,
	SINT1:      1,
	MONEY:      8,
	SHORTDATE:  4,
	SHORTMONEY: 4,
	TIME:       4,
	UINT2:      2,
	UINT4:      4,
	UINT8:      8,
}

// ByteSize returns the length of the DataTypes byte representation.
// Non-fixed-length and nullable data types return -1.
func (t DataType) ByteSize() int {
	byteSize, ok := ByteSizes[t]
	if !ok {
		return -1
	}
	return byteSize
}
