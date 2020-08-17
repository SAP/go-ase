// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package types

// LengthBytes maps a DataType to number of bytes in their length
// property.
// Only non-fixed-length or nullable DataTypes are listed.
var LengthBytes = map[DataType]int{
	BIGDATETIMEN: 1,
	BIGTIMEN:     1,
	BINARY:       1,
	BOUNDARY:     1,
	CHAR:         1,
	DATEN:        1,
	DATETIMEN:    1,
	DECN:         1,
	FLTN:         1,
	IMAGE:        4,
	INTN:         1,
	LONGBINARY:   4,
	LONGCHAR:     4,
	MONEYN:       1,
	NUMN:         1,
	SENSITIVITY:  1,
	TEXT:         4,
	TIMEN:        1,
	UINTN:        1,
	UNITEXT:      4,
	VARBINARY:    1,
	VARCHAR:      1,
	XML:          4,
}

// LengthBytes returns the number of bytes in the length property of
// the data type.
// For fixed-length data types -1 is returned.
func (t DataType) LengthBytes() int {
	lengthBytes, ok := LengthBytes[t]
	if !ok {
		return -1
	}
	return lengthBytes
}
