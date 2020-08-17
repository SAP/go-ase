// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"reflect"
	"testing"
)

func TestNewVersionString(t *testing.T) {
	cases := map[string]*Version{
		"0.0.0.0":  &Version{major: 0, minor: 0, sp: 0, patch: 0},
		"99.1.0.4": &Version{major: 99, minor: 1, sp: 0, patch: 4},
	}

	for inputString, expectVersion := range cases {
		t.Run(inputString,
			func(t *testing.T) {
				recvVersion, err := NewVersionString(inputString)
				if err != nil {
					t.Errorf("NewVersionString errored: %v", err)
					return
				}

				if !reflect.DeepEqual(recvVersion, expectVersion) {
					t.Errorf("Received Version does not match expected Version:")
					t.Errorf("Expected: %v", expectVersion)
					t.Errorf("Received: %v", recvVersion)
				}
			},
		)
	}
}

func TestVersion_Compare(t *testing.T) {
	cases := map[string]struct {
		a, b   Version
		expect int
	}{
		"equal": {
			a:      Version{major: 0, minor: 1, sp: 2, patch: 3},
			b:      Version{major: 0, minor: 1, sp: 2, patch: 3},
			expect: 0,
		},
		"a lower than b": {
			a:      Version{major: 0, minor: 1, sp: 6, patch: 5},
			b:      Version{major: 1, minor: 5, sp: 3, patch: 6},
			expect: -1,
		},
		"a higher than b": {
			a:      Version{major: 0, minor: 6, sp: 6, patch: 5},
			b:      Version{major: 0, minor: 5, sp: 3, patch: 6},
			expect: 1,
		},
	}

	for title, cas := range cases {
		t.Run(title,
			func(t *testing.T) {
				recv := cas.a.Compare(cas.b)
				if recv != cas.expect {
					t.Errorf("Comparing %v and %v should return %d, returned %d instead",
						cas.a, cas.b, cas.expect, recv)
				}
			},
		)
	}
}
