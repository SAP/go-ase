package capability

type Version interface {
	VersionString() string
	SetCapability(*Capability, bool)
	Has(*Capability) bool
}

// Version represent a single version of a Target.
type DefaultVersion struct {
	// Spec is the string representation of the version.
	// Depending on the used versioning scheme these may vary wildly.
	spec         string
	capabilities map[*Capability]bool
}

func NewVersion(spec string) Version {
	return &DefaultVersion{spec, map[*Capability]bool{}}
}

func (v DefaultVersion) VersionString() string {
	return v.spec
}

func (v *DefaultVersion) SetCapability(cap *Capability, b bool) {
	v.capabilities[cap] = b
}

func (v DefaultVersion) Has(cap *Capability) bool {
	canCap, ok := v.capabilities[cap]
	if !ok {
		return false
	}
	return canCap
}
