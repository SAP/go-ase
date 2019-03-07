package capability

import "fmt"

// Capability is a representation for a defined behaviour of an
// application, API, library, etc.pp.
type Capability struct {
	// The description is optional and not used by this package.
	Description string
	// Ranges in which the Capability exists. See main_test.go for more
	// elaborate examples.
	VersionRanges []VersionRange
}

// NewCapability creates a new Capability instance with the description
// and ranges provided.
//
// Versions are read in pairs and considered the lower and upper bound
// of a range respectively. If the last passed version does not have
// a partner the upper bound is set as empty.
//
// Examples:
//   versionRanges: 0.1.0, 0.2.0, 0.5.0, 1.0.0
//   Result:
//     - 0.1.0 -> 0.2.0
//     - 0.5.0 -> 1.0.0
//   versionRanges: 1.0.0
//   Result:
//     - 1.0.0 -> (no upper bound)
//   versionRanges: 1.0.0, ""
//   Result:
//     - 1.0.0 -> (no upper bound)
//   versionRanges: "", 5.1.0, 1.0.0, 2.0.0, 3.5.0
//   Result:
//     - (no lower bound) -> 5.1.0
//     - 1.0.0 -> 2.0.0
//     - 3.5.0 -> (no upper bound)
func NewCapability(description string, versionRanges ...string) *Capability {
	c := &Capability{
		Description:   description,
		VersionRanges: []VersionRange{},
	}

	curRange := VersionRange{}
	for i, s := range versionRanges {
		if i%2 == 0 {
			curRange.Introduced = s
			continue
		}

		curRange.Removed = s
		c.VersionRanges = append(c.VersionRanges, curRange)
		curRange = VersionRange{}
	}

	if curRange.Introduced != "" {
		c.VersionRanges = append(c.VersionRanges, curRange)
	}

	return c
}

func (cap Capability) String() string {
	s := fmt.Sprintf("Capability %s -> (", cap.Description)
	for i, vrange := range cap.VersionRanges {
		s += vrange.String()
		if i+1 < len(cap.VersionRanges) {
			s += ", "
		}
	}
	s += ")"

	return s
}
