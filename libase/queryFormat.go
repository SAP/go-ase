package libase

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// QueryFormat replaces each question mark in `query` with the value at
// the respective index.
//
// Example:
//	query: `select * from table where name = ?`
//	values: aName
//	-> `select * from table where name = "aName"`
func QueryFormat(query string, values ...driver.Value) (string, error) {
	if strings.Count(query, "?") != len(values) {
		return "", fmt.Errorf("Number of placeholders and passed arguments does not match. Placeholders: %d, Arguments: %d",
			strings.Count(query, "?"), len(values))
	}

	if len(values) == 0 {
		return query, nil
	}

	pass := make([]interface{}, len(values))
	for i, val := range values {
		pass[i] = val
	}

	return fmt.Sprintf(strings.Replace(query, "?", "%v", -1), pass...), nil
}

// TODO: NamedValue.Name should be used as a parameter identifier.
// TODO: Support named parameters
func NamedQueryFormat(query string, values ...driver.NamedValue) (string, error) {
	convertedValues := make([]driver.Value, len(values))
	for i, value := range values {
		convertedValues[i] = value.Value
	}

	return QueryFormat(query, convertedValues...)
}
