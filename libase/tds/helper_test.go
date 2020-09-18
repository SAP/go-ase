// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"reflect"
	"testing"
)

func TestDeBitmask(t *testing.T) {
	cases := map[string]struct {
		bitmask, maxValue int
		expected          []int
	}{
		"simple binary": {
			bitmask:  0b00010110,
			maxValue: 255,
			expected: []int{
				0b00000010,
				0b00000100,
				0b00010000,
			},
		},
		"simple hex": {
			bitmask:  0x41,
			maxValue: 0x80,
			expected: []int{
				0x1,
				0x40,
			},
		},
		"maxima binary": {
			bitmask:  0b10010110,
			maxValue: 128,
			expected: []int{
				0b00000010,
				0b00000100,
				0b00010000,
				0b10000000,
			},
		},
		"maxima hex": {
			bitmask:  0xc1,
			maxValue: 0x80,
			expected: []int{
				0x1,
				0x40,
				0x80,
			},
		},
	}

	for title, cas := range cases {
		t.Run(title,
			func(t *testing.T) {
				recv := deBitmask(cas.bitmask, cas.maxValue)
				if !reflect.DeepEqual(recv, cas.expected) {
					t.Errorf("Received unexpected response:")
					t.Errorf("Expected: %#v", cas.expected)
					t.Errorf("Received: %#v", recv)
				}
			},
		)
	}
}

type deBitmaskTypeUint uint

const (
	deBitmaskTypeUint1 deBitmaskTypeUint = 0x1
	deBitmaskTypeUint2 deBitmaskTypeUint = 0x2
	deBitmaskTypeUint3 deBitmaskTypeUint = 0x4
	deBitmaskTypeUint4 deBitmaskTypeUint = 0x8
)

func (t deBitmaskTypeUint) String() string {
	switch t {
	case deBitmaskTypeUint1:
		return "deBitmaskTypeUint1"
	case deBitmaskTypeUint2:
		return "deBitmaskTypeUint2"
	case deBitmaskTypeUint3:
		return "deBitmaskTypeUint3"
	case deBitmaskTypeUint4:
		return "deBitmaskTypeUint4"
	default:
		return ""
	}
}

type deBitmaskTypeInt int

const (
	deBitmaskTypeInt1 deBitmaskTypeInt = 0x1
	deBitmaskTypeInt2 deBitmaskTypeInt = 0x2
	deBitmaskTypeInt3 deBitmaskTypeInt = 0x8
	deBitmaskTypeInt4 deBitmaskTypeInt = 0x10
)

func (t deBitmaskTypeInt) String() string {
	switch t {
	case deBitmaskTypeInt1:
		return "deBitmaskTypeInt1"
	case deBitmaskTypeInt2:
		return "deBitmaskTypeInt2"
	case deBitmaskTypeInt3:
		return "deBitmaskTypeInt3"
	case deBitmaskTypeInt4:
		return "deBitmaskTypeInt4"
	default:
		return ""
	}
}

func TestDeBitmaskString(t *testing.T) {
	cases := map[string]struct {
		bitmask, maxValue int
		defaultValue      string
		stringerFn        func(int) string
		expect            string
	}{
		"uint": {
			bitmask:    int(deBitmaskTypeUint1 | deBitmaskTypeUint3),
			maxValue:   int(deBitmaskTypeUint4),
			stringerFn: func(i int) string { return deBitmaskTypeUint(i).String() },
			expect:     "deBitmaskTypeUint1|deBitmaskTypeUint3",
		},
		"uint default": {
			bitmask:      0,
			maxValue:     int(deBitmaskTypeUint4),
			defaultValue: "unused",
			stringerFn:   func(i int) string { return deBitmaskTypeUint(i).String() },
			expect:       "unused",
		},
		"int": {
			bitmask:    int(deBitmaskTypeInt2 | deBitmaskTypeInt4),
			maxValue:   int(deBitmaskTypeInt4),
			stringerFn: func(i int) string { return deBitmaskTypeInt(i).String() },
			expect:     "deBitmaskTypeInt2|deBitmaskTypeInt4",
		},
		"int default": {
			bitmask:      0,
			maxValue:     int(deBitmaskTypeInt4),
			defaultValue: "default",
			stringerFn:   func(i int) string { return deBitmaskTypeInt(i).String() },
			expect:       "default",
		},
	}

	for title, cas := range cases {
		t.Run(title,
			func(t *testing.T) {
				result := deBitmaskString(cas.bitmask, cas.maxValue, cas.stringerFn, cas.defaultValue)
				if result != cas.expect {
					t.Errorf("Received unexpected result:")
					t.Errorf("Expected: %s", cas.expect)
					t.Errorf("Received: %s", result)
				}
			},
		)
	}
}
