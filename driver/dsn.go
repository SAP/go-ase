package driver

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// DsnInfo represents all required information to open a connection to
// an ASE server.
type DsnInfo struct {
	Host, Port, Username, Password string
	ConnectProps                   url.Values
}

// ParseDSN parses a DSN into a DsnInfo struct.
//
// Accepted DSNs are either in URI or simple form:
// URI: ase://user:pass@host:port?key=val
// Simple: username=user password=password host=host port=port key=val
//
// To use special characters in your DSN use the simple form.
//
// When using the simple form values containing whitespaces must be
// quoted with double or single quotation marks.
//		username=user password="a password" host=host port=port
//		username=user password='a password' host=host port=port
func ParseDSN(dsn string) (*DsnInfo, error) {
	var dsnInfo *DsnInfo
	var err error

	if strings.HasPrefix(dsn, "ase:/") {
		dsnInfo, err = parseDsnUri(dsn)
	} else {
		dsnInfo, err = parseDsnSimple(dsn)
	}

	if err != nil {
		return nil, err
	}

	missingFields := []string{}

	checkFields := map[string]string{
		"Host":     dsnInfo.Host,
		"Port":     dsnInfo.Port,
		"Username": dsnInfo.Username,
		"Password": dsnInfo.Password,
	}

	for field, value := range checkFields {
		if value == "" {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		sort.Strings(missingFields)
		return nil, fmt.Errorf("Missing fields: %s", strings.Join(missingFields, ", "))
	}

	return dsnInfo, nil
}

// parseDsnUri parses a DSN in URI form and returns the resulting
// DsnInfo without checking for missing values.
func parseDsnUri(dsn string) (*DsnInfo, error) {
	url, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN: %v", err)
	}

	userName := ""
	password := ""
	if url.User != nil {
		userName = url.User.Username()
		password, _ = url.User.Password()
	}

	return &DsnInfo{
		Host:         url.Hostname(),
		Port:         url.Port(),
		Username:     userName,
		Password:     password,
		ConnectProps: url.Query(),
	}, nil
}

// parseDsnSimple parses a DSN in the simple form and returns the
// resulting DsnInfo without checking for missing values.
func parseDsnSimple(dsn string) (*DsnInfo, error) {
	dsni := &DsnInfo{
		ConnectProps: url.Values{},
	}

	// Valid quotation marks to detect values with whitespaces
	quotations := []byte{'\'', '"'}

	// Split the DSN on whitespace - any quoted values containing
	// whitespaces will be concatenated in the first step in the loop.
	dsnS := strings.Split(dsn, " ")
	for len(dsnS) > 0 {
		var part string
		part, dsnS = dsnS[0], dsnS[1:]

		// If the value starts with a quotation mark consume more parts
		// until the quotation is finished.
		for _, quot := range quotations {
			if !strings.Contains(part, "="+string(quot)) {
				continue
			}

			for part[len(part)-1] != quot {
				part = strings.Join([]string{part, dsnS[0]}, " ")
				dsnS = dsnS[1:]
			}
			break
		}

		partS := strings.SplitN(part, "=", 2)
		if len(partS) != 2 {
			return nil, fmt.Errorf("Recognized DSN part does not contain key/value parts: %s", partS)
		}

		key, value := partS[0], partS[1]

		// Remove quotation from value
		for _, quot := range quotations {
			if value[0] == quot && value[len(value)-1] == quot {
				value = value[1 : len(value)-1]
			}
		}

		switch key {
		case "host":
			fallthrough
		case "hostname":
			dsni.Host = value
		case "port":
			dsni.Port = value
		case "user":
			fallthrough
		case "username":
			dsni.Username = value
		case "pass":
			fallthrough
		case "password":
			dsni.Password = value
		default:
			dsni.ConnectProps.Add(key, value)
		}
	}

	return dsni, nil
}
