package libase

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// DsnInfo represents all required information to open a connection to
// an ASE server.
type DsnInfo struct {
	Host, Port, Username, Password, Userstorekey, Database string
	ConnectProps                                           url.Values
}

func (dsnInfo DsnInfo) AsSimple() string {
	ret := []string{}

	fields := map[string]string{
		"host":         dsnInfo.Host,
		"port":         dsnInfo.Port,
		"username":     dsnInfo.Username,
		"password":     dsnInfo.Password,
		"userstorekey": dsnInfo.Userstorekey,
		"database":     dsnInfo.Database,
	}

	for _, key := range []string{"username", "password", "host", "port", "userstorekey", "database"} {
		value := fields[key]
		if value != "" {
			ret = append(ret, fmt.Sprintf("%s='%s'", key, value))
		}
	}

	for key, valueL := range dsnInfo.ConnectProps {
		for _, value := range valueL {
			ret = append(ret, fmt.Sprintf("%s='%s'", key, value))
		}
	}

	return strings.Join(ret, " ")
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

	// Check ubiquitous properties
	fields := map[string]string{
		"Host": dsnInfo.Host,
		"Port": dsnInfo.Port,
	}
	if err := checkFields(*dsnInfo, fields); err != nil {
		return nil, err
	}

	// Check that any of the authentication methods are set
	// This would also be caught by the checks further down, but this
	// explicitly tells the user that either the userstorekey _or_
	// a username/password combination can be used.
	if len(dsnInfo.Userstorekey) == 0 && len(dsnInfo.Username) == 0 && len(dsnInfo.Password) == 0 {
		return nil, fmt.Errorf("Either userstorekey or username and password must be set")
	}

	// Exit early if userstorekey is set and ignore username/password
	// properties
	if len(dsnInfo.Userstorekey) > 0 {
		return dsnInfo, nil
	}

	// Check username/password
	fields = map[string]string{
		"Username": dsnInfo.Username,
		"Password": dsnInfo.Password,
	}
	if err := checkFields(*dsnInfo, fields); err != nil {
		return nil, err
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
		if value != "" {
			for _, quot := range quotations {
				if value[0] == quot && value[len(value)-1] == quot {
					value = value[1 : len(value)-1]
				}
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
		case "userstorekey":
			dsni.Userstorekey = value
		case "database":
			dsni.Database = value
		default:
			dsni.ConnectProps.Add(key, value)
		}
	}

	return dsni, nil
}

func checkFields(dsn DsnInfo, fields map[string]string) error {
	missingFields := []string{}

	for field, value := range fields {
		if value == "" {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) == 0 {
		return nil
	}

	sort.Strings(missingFields)
	return fmt.Errorf("Missing fields: %s", strings.Join(missingFields, ", "))
}
