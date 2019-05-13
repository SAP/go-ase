package capability

import (
	"fmt"
	"testing"
)

func TestVersionCompareSemantic(t *testing.T) {
	cases := map[string]struct {
		a, b   string
		result int
		err    error
	}{
		"less": {
			a:      "0.1.0",
			b:      "0.2.0",
			result: -1,
			err:    nil,
		},
		"equal": {
			a:      "1.0.0",
			b:      "1.0.0",
			result: 0,
			err:    nil,
		},
		"greater": {
			a:      "2.1.2",
			b:      "0.5.1",
			result: 1,
			err:    nil,
		},
		"err a": {
			a:      "",
			b:      "random text",
			result: 0,
			err:    fmt.Errorf("Malformed version: "),
		},
		"err b": {
			a:      "0.1",
			b:      "random text",
			result: 0,
			err:    fmt.Errorf("Malformed version: random text"),
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				i, err := VersionCompareSemantic(cas.a, cas.b)
				if cas.err != nil {
					if err == nil {
						t.Errorf("Expected error, got nil")
						return
					}

					if err.Error() != cas.err.Error() {
						t.Errorf("Received expected error with unexpected message")
						t.Errorf("Expected: %v", cas.err)
						t.Errorf("Received: %v", err)
						return
					}
				} else if err != nil {
					t.Errorf("Received unexpected error: %v", err)
					return
				}

				if i != cas.result {
					t.Errorf("Result '%d' does not match expected '%d'", i, cas.result)
				}
			},
		)
	}
}
