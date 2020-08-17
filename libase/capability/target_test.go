// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package capability

import (
	"fmt"
	"os"
	"testing"
)

func ExampleTarget_Version() {
	cap1 := NewCapability("cap1", "1.0.0")
	cap2 := NewCapability("cap2", "0.9.5", "1.5.0")
	t := Target{nil, []*Capability{cap1, cap2}}

	v1, err := t.Version("1.0.1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred retrieving Version 1.0.1: %v", err)
		return
	}

	v2, err := t.Version("0.9.5")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred retrieving Version 0.9.5: %v", err)
		return
	}

	for _, version := range []Version{v1, v2} {
		fmt.Printf("Version: %s, Capability cap1: %t, Capability cap2: %t\n",
			version.VersionString(), version.Has(cap1), version.Has(cap2))
	}
	// Output:
	// Version: 1.0.1, Capability cap1: true, Capability cap2: true
	// Version: 0.9.5, Capability cap1: false, Capability cap2: true
}

func TestTarget_Version(t *testing.T) {
	spec := "1.1.0"
	capsEnabled := []*Capability{Feature1, Feature2, Feature3, Bugfix1, Bugfix2}
	capsDisabled := []*Capability{Bugfix3}

	v, err := AppTarget.Version(spec)
	if err != nil {
		t.Errorf("Received unexpected error: %v", err)
		return
	}

	if v == nil {
		t.Errorf("Received no error and no version")
		return
	}

	if v.VersionString() != spec {
		t.Errorf("Received version with different spec")
		t.Errorf("Expected: %s", spec)
		t.Errorf("Received: %s", v.VersionString())
	}

	for _, cap := range capsEnabled {
		if v.Has(cap) != true {
			t.Errorf("Capability for version '%s' is not enabled but should be: %s", v, cap)
		}
	}

	for _, cap := range capsDisabled {
		if v.Has(cap) != false {
			t.Errorf("Capability for version '%s' is enabled but shouldn't be: %s", v, cap)
		}
	}
}

func TestTarget_VersionErr(t *testing.T) {
	_, err := AppTarget.Version("1 1 0")
	if err == nil {
		t.Errorf("Expected error, got none")
		return
	}

	errMsg := "Received error comparing '0.5.0' against '1 1 0': Malformed version: 1 1 0"
	if err.Error() != errMsg {
		t.Errorf("Received error with unexpected message:")
		t.Errorf("Expected: %s", errMsg)
		t.Errorf("Received: %v", err)
	}
}
