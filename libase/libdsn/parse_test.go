package libdsn

import (
	"net/url"
	"reflect"
	"testing"

	validator "gopkg.in/go-playground/validator.v9"
)

func TestParseDsnUri(t *testing.T) {
	cases := map[string]struct {
		dsn     string
		info *Info
	}{
		"URI DSN": {
			dsn: "ase://user:password@fully.qualified.domain.name:4901?",
			info: &Info{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"URI DSN Hostname": {
			dsn: "ase://user:password@hostname:4901?",
			info: &Info{
				Host:         "hostname",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"URI DSN Properties": {
			dsn: "ase://user:password@hostname:4901?foo=bar&bar=baz&bar=baf",
			info: &Info{
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

				if !reflect.DeepEqual(res, cas.info) {
					t.Errorf("Received invalid parsed Info")
					t.Errorf("Expected: %+v", cas.info)
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
			dsn:      "ase://user:pass$@!%=word@hostname:4901?",
			errorMsg: "Failed to parse DSN using url.Parse: parse \"ase://user:pass$@!%=word@hostname:4901?\": invalid URL escape \"%=w\"",
		},
		"URI DSN Password with pound": {
			dsn:      "ase://user:pass$#@!%=word@hostname:4901?",
			errorMsg: "Failed to parse DSN using url.Parse: parse \"ase://user:pass$\": invalid port \":pass$\" after host",
		},
	}

	for name, cas := range cases {
		t.Run(name,
			func(t *testing.T) {
				res, err := parseDsnUri(cas.dsn)

				if err == nil {
					t.Errorf("Expected error, received nil")
				} else if err.Error() != cas.errorMsg {
					t.Errorf("Received invalid error message")
					t.Errorf("Expected: %s", cas.errorMsg)
					t.Errorf("Received: %s", err.Error())
				}

				if res != nil {
					t.Errorf("Received parsed Info %v, expected error: %s", res, cas.errorMsg)
				}
			},
		)
	}
}

func TestParseDsnSimple(t *testing.T) {
	cases := map[string]struct {
		dsn     string
		info *Info
	}{
		"Simple DSN": {
			dsn: "username=user password=\"password\" host=fully.qualified.domain.name port=4901",
			info: &Info{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"Simple DSN Hostname": {
			dsn: "username='user' password=password host=hostname port=4901",
			info: &Info{
				Host:         "hostname",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"Simple DSN Properties": {
			dsn: "username=user password=password host=hostname port=4901 foo=bar bar=baz bar=baf",
			info: &Info{
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
			info: &Info{
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

				if !reflect.DeepEqual(res, cas.info) {
					t.Errorf("Received invalid parsed Info")
					t.Errorf("Expected: %+v", cas.info)
					t.Errorf("Received: %+v", res)
				}
			},
		)
	}
}

func TestParseDSN(t *testing.T) {
	cases := map[string]struct {
		dsn     string
		info *Info
	}{
		"URI DSN": {
			dsn: "ase://user:password@fully.qualified.domain.name:4901?",
			info: &Info{
				Host:         "fully.qualified.domain.name",
				Port:         "4901",
				Username:     "user",
				Password:     "password",
				ConnectProps: url.Values{},
			},
		},
		"Simple DSN": {
			dsn: "username=user password=\"password\" host=fully.qualified.domain.name port=4901",
			info: &Info{
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

				if !reflect.DeepEqual(res, cas.info) {
					t.Errorf("Received invalid parsed Info")
					t.Errorf("Expected: %+v", cas.info)
					t.Errorf("Received: %+v", res)
				}
			},
		)
	}
}

func TestParseDSNFail(t *testing.T) {
	type failedField struct {
		namespace, tag string
	}

	cases := map[string]struct {
		simpleDsn, uriDsn string
		failedFields      []failedField
	}{
		"DSN URI Missing host": {
			uriDsn:    "ase://user:pass@:4901?",
			simpleDsn: "username=user password=pass port=4901",
			failedFields: []failedField{
				{namespace: "Info.Host", tag: "required"},
			},
		},
		"DSN Simple Missing host and user": {
			uriDsn:    "ase://:pass@:4901?",
			simpleDsn: "password=pass port=4901",
			failedFields: []failedField{
				{namespace: "Info.Host", tag: "required"},
				{namespace: "Info.Username", tag: "required"},
			},
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
						validationErrs, ok := err.(validator.ValidationErrors)
						if !ok {
							t.Errorf("Received error other than validator.ValidationErrors: %v", err)
							continue
						}

						for i, fieldError := range validationErrs {
							if fieldError.Namespace() != cas.failedFields[i].namespace || fieldError.Tag() != cas.failedFields[i].tag {
								t.Errorf("validator.FieldError does not match expected error")
								t.Errorf("Expected: %s %s", cas.failedFields[i].namespace, cas.failedFields[i].tag)
								t.Errorf("Received: %s %s", fieldError.Namespace(), fieldError.Tag())
							}
						}
					}

					if res != nil {
						t.Errorf("Expected error, received parsed Info: %v", res)
					}
				}
			},
		)
	}
}
