package capability

// Version defines the interface of a single version of a target.
type Version interface {
	VersionString() string
	SetCapability(*Capability, bool)
	Has(*Capability) bool
}

// DefaultVersion is an example implementation for Version and can also
// be used for composition.
type DefaultVersion struct {
	// Spec is the string representation of the version.
	// Depending on the used versioning scheme these may vary wildly.
	spec         string
	capabilities map[*Capability]bool
}

// NewVersion returns an initialized DefaultVersion.
func NewVersion(spec string) Version {
	return &DefaultVersion{spec, map[*Capability]bool{}}
}

// VersionString returns the version spec.
func (v DefaultVersion) VersionString() string {
	return v.spec
}

// SetCapability sets the capability as enabled or disabled.
func (v *DefaultVersion) SetCapability(cap *Capability, b bool) {
	v.capabilities[cap] = b
}

// Has returns if the capability is enabled or disabled.
func (v DefaultVersion) Has(cap *Capability) bool {
	canCap, ok := v.capabilities[cap]
	if !ok {
		return false
	}
	return canCap
}
