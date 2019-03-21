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
func QueryFormat(query string, values ...interface{}) (string, error) {
	if strings.Count(query, "?") != len(values) {
		return "", fmt.Errorf("Number of placeholders and passed arguments does not match. Placeholders: %d, Arguments: %d",
			strings.Count(query, "?"), len(values))
	}

	if len(values) == 0 {
		return query, nil
	}

	for _, value := range values {
		switch value.(type) {
		case string:
			query = strings.Replace(query, "?", "%q", 1)
		default:
			query = strings.Replace(query, "?", "%v", 1)
		}
	}

	return fmt.Sprintf(query, values...), nil
}

// TODO: NamedValue.Name should be used as a parameter identifier.
// TODO: Support named parameters
func NamedQueryFormat(query string, values ...driver.NamedValue) (string, error) {
	convertedValues := make([]interface{}, len(values))
	for i, value := range values {
		convertedValues[i] = value.Value
	}

	return QueryFormat(query, convertedValues...)
}
