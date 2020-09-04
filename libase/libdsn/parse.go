// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package libdsn

import (
	"fmt"
	"net/url"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

// ParseDSN parses a DSN into a Info struct.
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
//
// The DSN is validated using the struct tags and validator.
// Validation errors from validator are returned as-is for further
// processing.
func ParseDSN(dsn string) (*Info, error) {
	var info *Info
	var err error

	// Parse DSN
	if strings.HasPrefix(dsn, "ase:/") {
		info, err = parseDsnUri(dsn)
	} else {
		info, err = parseDsnSimple(dsn)
	}
	if err != nil {
		return nil, err
	}

	var filterFn validator.FilterFunc = filterNoUserStoreKey
	if info.Userstorekey != "" {
		filterFn = filterUserStoreKey
	}

	v := validator.New()
	err = v.StructFiltered(info, filterFn)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// filterUserStoreKey is the validator.FilterFunc for a Info struct
// with Userstorekey set.
func filterUserStoreKey(ns []byte) bool {
	switch string(ns) {
	case "Info.Username":
		return true
	case "Info.Password":
		return true
	case "Info.Database":
		return true
	case "Info.Host":
		return true
	case "Info.Port":
		return true
	}
	return false
}

// filterNoUserStoreKey is the validator.FilterFunc for a Info struct
// with Userstorekey unset.
func filterNoUserStoreKey(ns []byte) bool {
	return string(ns) == "Info.Userstorekey"
}

// parseDsnUri parses a DSN in URI form and returns the resulting
// Info.
func parseDsnUri(dsn string) (*Info, error) {
	url, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse DSN using url.Parse: %v", err)
	}

	dsni := &Info{
		Host:         url.Hostname(),
		Port:         url.Port(),
		ConnectProps: url.Query(),
	}

	// Assume that `astring` in the DSN ase://astring@hostname/ is the
	// userstorekey. This is parsed as the username by url.Parse.
	if url.User != nil {
		username := url.User.Username()
		password, ok := url.User.Password()

		if ok {
			// ase://username:password@hostname/
			dsni.Username = username
			dsni.Password = password
		} else {
			// ase://userstorekey@hostname/
			dsni.Userstorekey = username
		}
	}

	// Check ConnectProps for any values that should be set in struct
	ttf := dsni.tagToField(true)
	for prop, values := range dsni.ConnectProps {
		// Skip if values is empty
		if len(values) == 0 {
			continue
		}

		if _, ok := ttf[prop]; !ok {
			// ConnectProp is not in struct
			continue
		}

		if err := dsni.SetField(prop, values[len(values)-1]); err != nil {
			return nil, fmt.Errorf("error setting field %s with value '%s': %w",
				prop, values[len(values)-1], err)
		}

		dsni.ConnectProps.Del(prop)
	}

	return dsni, nil
}

// parseDsnSimple parses a DSN in the simple form and returns the
// resulting Info without checking for missing values.
func parseDsnSimple(dsn string) (*Info, error) {
	dsni := &Info{
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

		if err := dsni.SetField(key, value); err != nil {
			return nil, fmt.Errorf("error setting value '%s' for field %s: %w", value, key, err)
		}
	}

	return dsni, nil
}
