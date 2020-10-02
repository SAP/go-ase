// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package libdsn

import (
	"net/url"
	"os"
	"reflect"
	"testing"
)

func setEnv(kv map[string]string) (func(), error) {
	for key, value := range kv {
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
	}

	return func() {
		for key := range kv {
			os.Unsetenv(key)
		}
	}, nil
}

func TestNewInfoFromEnv(t *testing.T) {
	cases := map[string]struct {
		prefix   string
		env      map[string]string
		expected Info
	}{
		"no prefix": {
			prefix: "",
			env: map[string]string{
				"ASE_HOST":                "testhost",
				"ASE_PORT":                "4901",
				"ASE_USER":                "username",
				"ASE_PASS":                "password",
				"ASE_TLS_HOSTNAME":        "not-testhost",
				"ASE_TLS_SKIP_VALIDATION": "true",
				"ASE_TLS_CA":              "/tmp",
			},
			expected: Info{
				Host:              "testhost",
				Port:              "4901",
				Username:          "username",
				Password:          "password",
				Database:          "",
				PacketReadTimeout: 50,
				TLSHostname:       "not-testhost",
				TLSSkipValidation: true,
				TLSCAFile:         "/tmp",
				ConnectProps:      url.Values{},
			},
		},
		"prefix": {
			prefix: "NOTASE",
			env: map[string]string{
				"NOTASE_HOST":         "testhost",
				"NOTASE_PORT":         "4901",
				"NOTASE_USER":         "username",
				"NOTASE_USERSTOREKEY": "sapsa",
				"NOTASE_DATABASE":     "testdatabase",
			},
			expected: Info{
				Host:              "testhost",
				Port:              "4901",
				Username:          "username",
				Password:          "",
				Userstorekey:      "sapsa",
				Database:          "testdatabase",
				PacketReadTimeout: 50,
				ConnectProps:      url.Values{},
			},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				os.Clearenv()

				fn, err := setEnv(cas.env)
				if err != nil {
					t.Errorf("Error preparing environment: %v", err)
					return
				}
				defer fn()

				d, err := NewInfoFromEnv(cas.prefix)
				if err != nil {
					t.Errorf("Received unexpected error: %v", err)
					return
				}

				if !reflect.DeepEqual(cas.expected, *d) {
					t.Errorf("Received Info does not match expected:")
					t.Errorf("Expected: %#v", cas.expected)
					t.Errorf("Received: %#v", *d)
				}
			},
		)
	}
}

func TestInfo_tagToField(t *testing.T) {
	dsn := Info{
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

func TestInfo_AsSimple(t *testing.T) {
	cases := map[string]struct {
		dsn      Info
		expected string
	}{
		"Only required information": {
			dsn: Info{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "passwd",
			},
			expected: "host='hostname' password='passwd' port='4901' username='user'",
		},
		"Everything": {
			dsn: Info{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "passwd",
				Database: "db_example",
				ConnectProps: url.Values{
					"foo": []string{"bar"},
					"bar": []string{"", "baz"},
					"baz": []string{""},
				},
			},
			expected: "database='db_example' host='hostname' password='passwd' port='4901' username='user' bar='baz' baz='' foo='bar'",
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
