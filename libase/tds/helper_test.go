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
