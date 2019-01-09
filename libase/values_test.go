package libase

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

func TestValuesToNamedValues(t *testing.T) {
	cases := map[string]struct {
		input  []driver.Value
		output []driver.NamedValue
	}{
		"empty": {
			input:  []driver.Value{},
			output: []driver.NamedValue{},
		},
		"single": {
			input: []driver.Value{int(0)},
			output: []driver.NamedValue{
				driver.NamedValue{Name: "", Ordinal: 1, Value: 0},
			},
		},
		"mixed": {
			input: []driver.Value{int(0), "string"},
			output: []driver.NamedValue{
				driver.NamedValue{Name: "", Ordinal: 1, Value: 0},
				driver.NamedValue{Name: "", Ordinal: 2, Value: "string"},
			},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				result := ValuesToNamedValues(cas.input)

				if !reflect.DeepEqual(result, cas.output) {
					t.Errorf("Received invalid result")
					t.Errorf("Expected: %v", cas.output)
					t.Errorf("Received: %v", result)
				}
			},
		)
	}
}
