package dsn

import (
	"net/url"
	"testing"
)

func TestDsnInfo_tagToField(t *testing.T) {
	dsn := DsnInfo{
		Host:     "hostname before",
		Port:     "port before",
		Username: "user before",
		Password: "pass before",
	}

	ttf := dsn.tagToField(true)

	ttf["hostname"].SetString("hostname after")

	if dsn.Host != "hostname after" {
		t.Errorf("Unexpected value in dsn.Host: %v", dsn.Host)
	}

	ttf["host"].SetString("hostname after different key")
	if dsn.Host != "hostname after different key" {
		t.Errorf("Unexpected value in dsn.Host after setting through different key: %v", dsn.Host)
	}
}

func TestDsnInfo_AsSimple(t *testing.T) {
	cases := map[string]struct {
		dsn      DsnInfo
		expected string
	}{
		"Only required information": {
			dsn: DsnInfo{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "passwd",
			},
			expected: "host='hostname' pass='passwd' port='4901' user='user'",
		},
		"Everything": {
			dsn: DsnInfo{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "passwd",
				Database: "db_example",
				ConnectProps: url.Values{
					"foo": []string{"bar"},
					"bar": []string{"", "baz"},
				},
			},
			expected: "bar='' bar='baz' db='db_example' foo='bar' host='hostname' pass='passwd' port='4901' user='user'",
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				result := cas.dsn.AsSimple()
				if result != cas.expected {
					t.Errorf("Received invalid simple URI")
					t.Errorf("Expected: %s", cas.expected)
					t.Errorf("Received: %s", result)
				}
			},
		)
	}
}
