// This file serves as both setup for testing as well as a usage
// example.
package capability

import (
	"fmt"
	"os"
)

var (
	// Feature implemented in 0.5.0, not removed
	Feature1 = NewCapability("feature1", "0.5.0", "")
	// Feature implemented in 0.6.0, removed in 0.9.0 and reintroduced
	// in 1.1.0
	Feature2 = NewCapability("feature2", "0.6.0", "0.9.0", "1.1.0", "")
	// Feature implemented in 1.1.0
	Feature3 = NewCapability("feature3", "0.1.0", "0.2.0", "1.1.0")

	// Bug introduced before 0.8.0, bugfix implemented in 0.8.0
	Bugfix1 = &Capability{"bugfix1", []VersionRange{{"0.8.0", ""}}}
	// Bug introduced before 0.1.0, fix implemented in 0.1.0
	// Bug reappeared in 0.7.0, new fix implemented in 1.0.0
	Bugfix2 = &Capability{"bugfix2", []VersionRange{{"0.1.0", "0.7.0"}, {"1.0.0", ""}}}
	// Empty as bug exists but no bugfix has been published
	Bugfix3 = &Capability{"bugfix3", []VersionRange{}}

	AppTarget = &Target{
		nil,
		[]*Capability{Feature1, Feature2, Feature3, Bugfix1, Bugfix2},
	}

	AppVersion Version
)

func init() {
	v, err := AppTarget.Version("1.0.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to retrieve version 1.0.0: %v", err)
	}

	AppVersion = v
}
