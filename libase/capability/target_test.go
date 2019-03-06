package capability

import "testing"

func TestTarget_Version(t *testing.T) {
	spec := "1.1.0"
	capsEnabled := []*Capability{Feature1, Feature2, Feature3, Bugfix1, Bugfix2}
	capsDisabled := []*Capability{Bugfix3}

	v, err := AppTarget.Version(spec)
	if err != nil {
		t.Errorf("Received unexpected error: %v", err)
		return
	}

	if v == nil {
		t.Errorf("Received no error and no version")
		return
	}

	if v.VersionString() != spec {
		t.Errorf("Received version with different spec")
		t.Errorf("Expected: %s", spec)
		t.Errorf("Received: %s", v.VersionString())
	}

	for _, cap := range capsEnabled {
		if v.Can(cap) != true {
			t.Errorf("Capability for version '%s' is not enabled but should be: %s", v, cap)
		}
	}

	for _, cap := range capsDisabled {
		if v.Can(cap) != false {
			t.Errorf("Capability for version '%s' is enabled but shouldn't be: %s", v, cap)
		}
	}
}

func TestTarget_VersionErr(t *testing.T) {
	_, err := AppTarget.Version("1 1 0")
	if err == nil {
		t.Errorf("Expected error, got none")
		return
	}

	errMsg := "Received error comparing '0.5.0' against '1 1 0': Malformed version: 1 1 0"
	if err.Error() != errMsg {
		t.Errorf("Received error with unexpected message:")
		t.Errorf("Expected: %s", errMsg)
		t.Errorf("Received: %v", err)
	}
}
