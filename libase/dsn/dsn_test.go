package dsn

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func setEnv(prefix string, kv map[string]string) (func(), error) {
	if prefix != "" {
		prefix += "_"
	}

	for key, value := range kv {
		err := os.Setenv(strings.ToUpper(prefix+key), value)
		if err != nil {
			return nil, err
		}
	}

	return func() {
		for key := range kv {
			os.Unsetenv(strings.ToUpper(prefix + key))
		}
	}, nil
}

func TestNewDsnInfoFromEnv(t *testing.T) {
	cases := map[string]struct {
		prefix   string
		env      map[string]string
		expected DsnInfo
	}{
		"no prefix": {
			prefix: "",
			env: map[string]string{
				"host": "testhost",
				"port": "4901",
				"user": "username",
				"pass": "password",
			},
			expected: DsnInfo{
				Host:         "testhost",
				Port:         "4901",
				Username:     "username",
				Password:     "password",
				Database:     "",
				ConnectProps: url.Values{},
			},
		},
		"prefix": {
			prefix: "NOTASE",
			env: map[string]string{
				"host":         "testhost",
				"port":         "4901",
				"user":         "username",
				"userstorekey": "sapsa",
				"database":     "testdatabase",
			},
			expected: DsnInfo{
				Host:         "testhost",
				Port:         "4901",
				Username:     "username",
				Password:     "",
				Userstorekey: "sapsa",
				Database:     "testdatabase",
				ConnectProps: url.Values{},
			},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				passPrefix := cas.prefix
				if passPrefix == "" {
					passPrefix = "ASE"
				}

				os.Clearenv()

				fn, err := setEnv(passPrefix, cas.env)
				if err != nil {
					t.Errorf("Error preparing environment: %v", err)
					return
				}
				defer fn()

				d := NewDsnInfoFromEnv(cas.prefix)

				if !reflect.DeepEqual(cas.expected, *d) {
					t.Errorf("Received DsnInfo does not match expected:")
					t.Errorf("Expected: %#v", cas.expected)
					t.Errorf("Received: %#v", *d)
				}
			},
		)
	}
}

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
