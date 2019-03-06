package capability

type Version interface {
	VersionString() string
	SetCapability(*Capability, bool)
	Can(*Capability) bool
}

// Version represent a single version of a Target.
type version struct {
	// Spec is the string representation of the version.
	// Depending on the used versioning scheme these may vary wildly.
	spec         string
	capabilities map[*Capability]bool
}

func newVersion(spec string) Version {
	return &version{spec, map[*Capability]bool{}}
}

func (v version) VersionString() string {
	return v.spec
}

func (v version) SetCapability(cap *Capability, b bool) {
	v.capabilities[cap] = b
}

func (v version) Can(cap *Capability) bool {
	canCap, ok := v.capabilities[cap]
	if !ok {
		return false
	}
	return canCap
}
