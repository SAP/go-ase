package capability

import "testing"

func TestVersionRange_contains(t *testing.T) {
	cases := map[string]struct {
		vrange          VersionRange
		inside, outside []string
	}{
		"no bound set": {
			VersionRange{},
			[]string{},
			[]string{"1.0.0"},
		},
		"lower and upper set": {
			VersionRange{"0.5.0", "1.0.0"},
			[]string{"0.5.0", "0.7.0", "0.8", "0.9.9"},
			[]string{"0.4.0", "0.4.9", "1.0.0", "1.0.1"},
		},
		"only upper set": {
			VersionRange{"", "1.0.0"},
			[]string{"0.9.0", "0.1.0", "0.0.1"},
			[]string{"1.0.0", "1.5.0"},
		},
		"only lower set": {
			VersionRange{"0.5.0", ""},
			[]string{"0.5.0", "0.6.0", "9.5.1"},
			[]string{"0.0.0", "0.1.0", "0.4.9"},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				for _, in := range cas.inside {
					b, err := cas.vrange.contains(VersionCompareSemantic, in)
					if err != nil {
						t.Errorf("Received unexpected error: %v", err)
						continue
					}

					if b != true {
						t.Errorf("Version '%s' should be inside '%s' but .contains returned false",
							in, cas.vrange)
					}
				}

				for _, out := range cas.outside {
					b, err := cas.vrange.contains(VersionCompareSemantic, out)
					if err != nil {
						t.Errorf("Received unexpected error: %v", err)
						continue
					}

					if b != false {
						t.Errorf("Version '%s' should be outside '%s' but .contains returned true",
							out, cas.vrange)
					}
				}
			},
		)
	}
}

func TestVersionRange_containsErr(t *testing.T) {
	cases := map[string]struct {
		vrange    VersionRange
		version   string
		errString string
	}{
		"lower invalid": {
			VersionRange{"scrambled", "1.0.0"},
			"0.5.0",
			"Received error comparing 'scrambled' against '0.5.0': Malformed version: scrambled",
		},
		"upper invalid": {
			VersionRange{"0.1.0", "scrambled"},
			"0.5.0",
			"Received error comparing '0.5.0' against 'scrambled': Malformed version: scrambled",
		},
		"version invalid": {
			VersionRange{"0.1.0", "1.0.0"},
			"scrambled",
			"Received error comparing '0.1.0' against 'scrambled': Malformed version: scrambled",
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				_, err := cas.vrange.contains(VersionCompareSemantic, cas.version)
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}

				if err.Error() != cas.errString {
					t.Errorf("Received unexpected error message")
					t.Errorf("Expected: %v", cas.errString)
					t.Errorf("Received: %v", err)
				}
			},
		)
	}
}
