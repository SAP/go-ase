package libase

import (
	"database/sql/driver"
	"testing"
)

func TestQueryFormat(t *testing.T) {
	cases := map[string]struct {
		query  string
		values []driver.Value
		result string
	}{
		"string": {
			query:  "select * from a where b like ?",
			values: []driver.Value{"c"},
			result: "select * from a where b like \"c\"",
		},
		"ints": {
			query:  "select * from aTable where x = ?",
			values: []driver.Value{10},
			result: "select * from aTable where x = 10",
		},
		"mixed": {
			query:  "select * from aTable where x like ? and y = ?",
			values: []driver.Value{"aString", 10},
			result: "select * from aTable where x like \"aString\" and y = 10",
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				result, err := QueryFormat(cas.query, cas.values...)
				if err != nil {
					t.Errorf("Error formatting query: %v", err)
					return
				}

				if result != cas.result {
					t.Errorf("Formatted query does not match expected result:")
					t.Errorf("Expected: %s", cas.result)
					t.Errorf("Received: %s", result)
				}
			},
		)
	}
}
