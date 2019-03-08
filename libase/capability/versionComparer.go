package capability

import (
	semver "github.com/hashicorp/go-version"
)

// VersionComparer compares two versions and returns the following
// values:
// -1 when a is lower than b
// 0 when a and b are equal
// 1 when a is higher than b
type VersionComparer func(a, b string) (int, error)

// VersionCompareSemantic is the default method used to compare
// versions and only support semantic versioning.
//
// The underlying library used to compare versions is
// github.com/hashicorp/go-version.
func VersionCompareSemantic(a, b string) (int, error) {
	vA, err := semver.NewVersion(a)
	if err != nil {
		return 0, err
	}

	vB, err := semver.NewVersion(b)
	if err != nil {
		return 0, err
	}

	return vA.Compare(vB), nil
}
