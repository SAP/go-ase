// Package capability provides abstract structures to keep a history of
// capabilities and features.
//
// Capabilities and Targets should be exported. Versions can be either
// composed into another struct or exported as well, depending on the
// use case.
//
// Code relying on a capability (e.g. a workaround for a missing
// capability or bug) can then check against the version:
//
//	package capsForLibA
//
//	import "github.com/SAP/go-ase/libase/capabilities"
//
//	var (
//		Bugfix1 = capabilities.Capability{"bugfix1", {"0.9.0", ""}}
//		Target = Target{nil, {Bugfix1}}
//	)
//
//	package main
//
//
//	import "some.path/capsForLibA"
//
//	func main() {
//		x := connectToServer()
//		version, _ := Target.Version(x.Version)
//
//		if version.Can(capsForLibA.Bugfix1) {
//			// Bugfix is applied, use normally
//		} else {
//			// Bugfix not yet applied, use workaround
//		}
//	}
package capability
