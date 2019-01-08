package libase

import (
	"net/url"
	"reflect"
	"testing"
)

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
			expected: "username='user' password='passwd' host='hostname' port='4901'",
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
			expected: "username='user' password='passwd' host='hostname' port='4901' database='db_example' foo='bar' bar='' bar='baz'",
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

func TestParseDsnUri(t *testing.T) {
	cases := map[string]struct {
		dsn     string
		dsnInfo *DsnInfo
	}{
		"URI DSN": {
			dsn: "ase://user:password@fully.qualified.domain.name:4901?",
			dsnInfo: &DsnInfo{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"URI DSN Hostname": {
			dsn: "ase://user:password@hostname:4901?",
			dsnInfo: &DsnInfo{
				Host:         "hostname",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"URI DSN Properties": {
			dsn: "ase://user:password@hostname:4901?foo=bar&bar=baz&bar=baf",
			dsnInfo: &DsnInfo{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "password",
				ConnectProps: url.Values{
					"foo": []string{"bar"},
					"bar": []string{"baz", "baf"},
				},
			},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				res, err := parseDsnUri(cas.dsn)

				if err != nil {
					t.Errorf("Could not parse valid DSN '%s': %v", cas.dsn, err)
					return
				}

				if !reflect.DeepEqual(res, cas.dsnInfo) {
					t.Errorf("Received invalid parsed DsnInfo")
					t.Errorf("Expected: %+v", cas.dsnInfo)
					t.Errorf("Received: %+v", res)
				}
			},
		)
	}
}

func TestParseDsnUriFail(t *testing.T) {
	cases := map[string]struct {
		dsn, errorMsg string
	}{
		"URI DSN Password special characters": {
			dsn:      "ase://user:pass$#@!%=word@hostname:4901?",
			errorMsg: "Failed to parse DSN: parse ase://user:pass$#@!%=word@hostname:4901?: invalid URL escape \"%=w\"",
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				res, err := parseDsnUri(cas.dsn)

				if err == nil {
					t.Errorf("Expected error, received nil")
				} else {
					if err.Error() != cas.errorMsg {
						t.Errorf("Received invalid error message")
						t.Errorf("Expected: %s", cas.errorMsg)
						t.Errorf("Received: %s", err.Error())
					}
				}

				if res != nil {
					t.Errorf("Received parsed DsnInfo, expected error: %v", res)
				}
			},
		)
	}
}

func TestParseDsnSimple(t *testing.T) {
	cases := map[string]struct {
		dsn     string
		dsnInfo *DsnInfo
	}{
		"Simple DSN": {
			dsn: "username=user password=\"password\" host=fully.qualified.domain.name port=4901",
			dsnInfo: &DsnInfo{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"Simple DSN Hostname": {
			dsn: "username='user' password=password host=hostname port=4901",
			dsnInfo: &DsnInfo{
				Host:         "hostname",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"Simple DSN Properties": {
			dsn: "username=user password=password host=hostname port=4901 foo=bar bar=baz bar=baf",
			dsnInfo: &DsnInfo{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "password",
				ConnectProps: url.Values{
					"foo": []string{"bar"},
					"bar": []string{"baz", "baf"},
				},
			},
		},
		"Simple DSN with empty value": {
			dsn: "username=user password=password host=hostname port=4901 database=\"\" foo=bar bar= bar=baf",
			dsnInfo: &DsnInfo{
				Host:     "hostname",
				Port:     "4901",
				Username: "user",
				Password: "password",
				Database: "",
				ConnectProps: url.Values{
					"foo": []string{"bar"},
					"bar": []string{"", "baf"},
				},
			},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				res, err := ParseDSN(cas.dsn)

				if err != nil {
					t.Errorf("Could not parse valid DSN '%s': %v", cas.dsn, err)
					return
				}

				if !reflect.DeepEqual(res, cas.dsnInfo) {
					t.Errorf("Received invalid parsed DsnInfo")
					t.Errorf("Expected: %+v", cas.dsnInfo)
					t.Errorf("Received: %+v", res)
				}
			},
		)
	}
}

func TestParseDSN(t *testing.T) {
	cases := map[string]struct {
		dsn     string
		dsnInfo *DsnInfo
	}{
		"URI DSN": {
			dsn: "ase://user:password@fully.qualified.domain.name:4901?",
			dsnInfo: &DsnInfo{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"Simple DSN": {
			dsn: "username=user password=\"password\" host=fully.qualified.domain.name port=4901",
			dsnInfo: &DsnInfo{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				res, err := ParseDSN(cas.dsn)

				if err != nil {
					t.Errorf("Could not parse valid DSN '%s': %v", cas.dsn, err)
					return
				}

				if !reflect.DeepEqual(res, cas.dsnInfo) {
					t.Errorf("Received invalid parsed DsnInfo")
					t.Errorf("Expected: %+v", cas.dsnInfo)
					t.Errorf("Received: %+v", res)
				}
			},
		)
	}
}

func TestParseDSNFail(t *testing.T) {
	cases := map[string]struct {
		simpleDsn, uriDsn, errorMsg string
	}{
		"DSN URI Missing host": {
			uriDsn:    "ase://user:pass@:4901?",
			simpleDsn: "username=user password=pass port=4901",
			errorMsg:  "Missing fields: Host",
		},
		"DSN Simple Missing host and user": {
			uriDsn:    "ase://:pass@:4901?",
			simpleDsn: "password=pass port=4901",
			errorMsg:  "Missing fields: Host, Username",
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				for _, dsn := range []string{cas.uriDsn, cas.simpleDsn} {
					res, err := ParseDSN(dsn)

					if err == nil {
						t.Errorf("Expected error, received nil")
					} else {
						if err.Error() != cas.errorMsg {
							t.Errorf("Received invalid error message")
							t.Errorf("Expected: %s", cas.errorMsg)
							t.Errorf("Received: %s", err.Error())
						}
					}

					if res != nil {
						t.Errorf("Received parsed DsnInfo, expected error: %v", res)
					}
				}
			},
		)
	}
}
