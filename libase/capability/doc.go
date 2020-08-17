// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

/*
Package capability provides abstract structures to keep a history of
capabilities and features.

Capabilities and Targets should be exported. Versions can be either
composed into another struct or exported as well, depending on the
use case.

Code relying on a capability (e.g. a workaround for a missing
capability or bug) can then check against the version:

In git.domain.tld/libA/libACaps:

	package libACaps

	import "github.com/SAP/go-ase/libase/capability"

	var (
		// Bug appeared in version 0.9.0, no published version has a fix
		Bug1 = capability.NewCapability("ticket #15", "0.9.0")
		Target = capability.Target{nil, {Bug1}}
	)

In git.domain.tld/module2/cmd/aCmd:

	package main

	import "git.domain.tld/libA/libACaps"

	func main() {
		x := connectToServer()
		version, _ := libACaps.Target.Version(x.Version)

		if version.Has(libACaps.Bug1) {
			// Bugfix not yet applied, use workaround
		} else {
			// Bugfix is applied, use normally
		}
	}

See defaultVersion_test.go for more elaborate and commented examples.
*/
package capability
