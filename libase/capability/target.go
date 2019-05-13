package capability

import "fmt"

// Target describes an application, API, library, etc.pp. and its
// capabilities.
type Target struct {
	// Function used to compare versions.
	VersionComparer VersionComparer
	// The list of registered capabilities.
	Capabilities []*Capability
}

// Version is the interface that allows a Target to set capabilities.
type Version interface {
	VersionString() string
	SetCapability(*Capability, bool)
	Has(*Capability) bool
}

// Version returns a new version and calls .SetCapabilities with it.
func (target Target) Version(spec string) (Version, error) {
	v := NewVersion(spec)

	err := target.SetCapabilities(v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// SetVersion enables and disabled capabilities on a Version based on
// its capabilities and their version ranges.
//
// An error is only returned in two cases:
// - VersionComparer cannot parse the version of Version.VersionString
//   or the version of a lower/upper bound in a VersionRange.
// - A VersionRange of a Capability has lower and upper bound set with
//   the lower bound being equal or greater than the upper bound.
func (target Target) SetCapabilities(v Version) error {
	cmpFn := target.VersionComparer
	if cmpFn == nil {
		cmpFn = VersionCompareSemantic
	}

	for _, cap := range target.Capabilities {
		for _, vrange := range cap.VersionRanges {
			if vrange.Introduced != "" && vrange.Removed != "" {
				i, err := cmpFn(vrange.Introduced, vrange.Removed)
				if err != nil {
					return fmt.Errorf("Failed to compare lower and upper bound of VersionRange %s in Capability %s: %v",
						vrange, cap, err)
				}
				if i >= 0 {
					return fmt.Errorf("VersionRange '%s' of Capability '%s' is invalid, lower bound is greater or equal to upper bound",
						vrange, cap)
				}
			}

			contains, err := vrange.contains(cmpFn, v.VersionString())
			if err != nil {
				return err
			}

			// Version is within a range of the capability, set
			// and continue to next capability
			if contains {
				v.SetCapability(cap, true)
				break
			}
			v.SetCapability(cap, false)
		}
	}

	return nil
}
