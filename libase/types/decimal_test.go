package types

import "testing"

func TestDecimal_SetString(t *testing.T) {
	testCases := map[string]struct {
		precision, scale int
		testString       string
		expectedErr      error
	}{
		"-0.0001":                {precision: 5, scale: 4},
		"0.0":                    {precision: 2, scale: 1},
		"1234.5678":              {precision: 8, scale: 4},
		"1234.05678":             {precision: 9, scale: 5},
		"1234567890123456789.0":  {precision: 38, scale: 19},
		"9999999999999999999.0":  {precision: 38, scale: 19},
		"-1234567890123456789.0": {precision: 38, scale: 19},
		"-9999999999999999999.0": {precision: 38, scale: 19},
		"0.1234567890123456789":  {precision: 38, scale: 19},
		"0.9999999999999999999":  {precision: 38, scale: 19},
		"-0.1234567890123456789": {precision: 38, scale: 19},
		"-0.9999999999999999999": {precision: 38, scale: 19},
	}

	for name, cas := range testCases {
		t.Run(name,
			func(t *testing.T) {
				dec, _ := NewDecimal(cas.precision, cas.scale)
				err := dec.SetString(name)
				if err != cas.expectedErr {
					t.Errorf("Received unexpected error:")
					t.Errorf("Expected: %w", cas.expectedErr)
					t.Errorf("Received: %w", err)
				}

				if dec.String() != name {
					t.Errorf("Received unexpected string:")
					t.Errorf("Expected: %s", name)
					t.Errorf("Received: %s", dec.String())
				}
			},
		)
	}
}
