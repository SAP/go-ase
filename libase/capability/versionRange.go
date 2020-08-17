// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package capability

import "fmt"

// VersionRange describes the range of versions a capability is
// available for.
//
// VersionIntroduced is an inclusive bound, describing the version in
// which the capability was introduced.
//
// VersionRemoved is an exclusive bound, describing the version in which
// the capability is not available anymore.
type VersionRange struct {
	Introduced, Removed string
}

func (vrange VersionRange) String() string {
	return fmt.Sprintf("'%s' -> '%s'", vrange.Introduced, vrange.Removed)
}

func (vrange VersionRange) contains(fn VersionComparer, version string) (bool, error) {
	// Range magically doesn't exist, Version can't be contained in
	// a non-existing range.
	if vrange.Introduced == "" && vrange.Removed == "" {
		return false, nil
	}

	var lower, upper int
	var err error

	// Compare lower inclusive bound if it is set
	if vrange.Introduced != "" {
		lower, err = fn(vrange.Introduced, version)
		if err != nil {
			return false, fmt.Errorf("Received error comparing '%s' against '%s': %v",
				vrange.Introduced, version, err)
		}

		// Lower bound is set, upper is not; only check lower bound
		if vrange.Removed == "" {
			return lower <= 0, nil
		}
	}

	// Compare upper exclusive bound if it is set
	if vrange.Removed != "" {
		upper, err = fn(version, vrange.Removed)
		if err != nil {
			return false, fmt.Errorf("Received error comparing '%s' against '%s': %v",
				version, vrange.Removed, err)
		}

		// Upper bound is set, lower is not; only check upper bound
		if vrange.Introduced == "" {
			return upper < 0, nil
		}
	}

	// Both bounds are set, check both
	return lower <= 0 && upper < 0, nil
}
