// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package namepool

import (
	"testing"
)

func TestNewPool(t *testing.T) {
	cases := map[string]struct {
		format string
		first  string
	}{
		"empty": {
			format: "",
			first:  "%!(EXTRA uint64=1)",
		},
		"only id": {
			format: "%d",
			first:  "1",
		},
		"format with only %d": {
			format: "name %d",
			first:  "name 1",
		},
		"format with multiple formats": {
			format: "name %d %d %s",
			first:  "name 1 %!d(MISSING) %!s(MISSING)",
		},
		"format with string verb": {
			format: "name %s",
			first:  "name %!s(uint64=1)",
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			pool := Pool(cas.format)
			name := pool.Acquire()
			defer pool.Release(name)

			if name.Name() != cas.first {
				t.Errorf("Expected to receive '%s' as first name, received: %s", cas.first, name.Name())
			}
		})
	}
}

func TestPool_Release(t *testing.T) {
	pool := Pool("%d")

	name := pool.Acquire()
	if name == nil {
		t.Errorf("Acquired Name is nil")
		return
	}

	pool.Release(name)
	if name.id != nil {
		t.Errorf("Released Name has non-nil ID pointer")
	}
	if (*name).name != "" {
		t.Errorf("Released Name has non-empty name")
	}
}
