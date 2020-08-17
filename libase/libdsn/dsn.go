// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// Package libdsn handles data source names.
package libdsn

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
)

// Info represents all required information to open a connection to
// an ASE server.
//
// The json tag is the expected string in a simple URI.
type Info struct {
	Host         string     `json:"host" multiref:"hostname" validate:"required"`
	Port         string     `json:"port" validate:"required"`
	Username     string     `json:"username" multiref:"user" validate:"required"`
	Password     string     `json:"password" multiref:"passwd,pass" validate:"required"`
	Userstorekey string     `json:"userstorekey" multiref:"key" validate:"required"`
	Database     string     `json:"database" multiref:"db"`
	ConnectProps url.Values `json:"connectprops"`
}

// NewInfo returns an initialized Info.
func NewInfo() *Info {
	dsn := &Info{}
	dsn.ConnectProps = url.Values{}
	return dsn
}

// NewInfoFromEnv returns a new Info and fills it with data from
// the environment.
//
// If prefix is empty it is set as `ASE`.
//
// Recognized environments variables are in the form of <prefix>_<json
// tag>. E.g. `.Host` with the prefix `""` would recognize `ASE_HOST`
// and `ASE_HOSTNAME`.
//
// Properties with dashes are recognized with double underscored
// instead.
// E.g. the property `cgo-callback-client` can be passed as
// `CGO__CALLBACK__CLIENT`.
func NewInfoFromEnv(prefix string) *Info {
	dsn := NewInfo()

	if prefix == "" {
		prefix = "ASE"
	}
	prefix += "_"

	ttf := dsn.tagToField(true)
	for _, env := range os.Environ() {
		envSplit := strings.SplitN(env, "=", 2)
		key, value := envSplit[0], envSplit[1]

		if !strings.HasPrefix(key, prefix) {
			continue
		}

		key = strings.ToLower(strings.TrimPrefix(key, prefix))
		key = strings.ReplaceAll(key, "__", "-")
		if field, ok := ttf[key]; ok {
			field.SetString(value)
		} else {
			dsn.ConnectProps.Add(key, value)
		}
	}

	return dsn
}

// tagToField returns a mapping from json metadata tags to
// reflect.Values.
// If multiref is true the metadata tags from `multiref` will also be
// mapped to their field.Value.
// If multiref is false only the json tag will be mapped.
// multiref = true:
//   map[string]reflect.Value{"host": info.Host, "hostname": info.Host}
// multiref = false:
//   map[string]reflect.Value{"host": info.Host}
func (info *Info) tagToField(multiref bool) map[string]reflect.Value {
	tTF := map[string]reflect.Value{}
	// The accepted type of ValueOf is interface, which still allows
	// accessing the metadata but not the fields, since an interface
	// doesn't have field.
	// By passing a pointer it is possible to call .Elem(), which
	// returns a reflect.Value representation of the passed struct
	// - which allows to access its fields.
	v := reflect.ValueOf(info).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		// Skip over ConnectProps, this member is handled specially
		// since it is a map with additional methods.
		if t.Field(i).Name == "ConnectProps" {
			continue
		}

		// Grab json tag
		names := strings.Split(t.Field(i).Tag.Get("json"), ",")
		names = []string{names[0]}
		if multiref {
			// Grab multiref tags if enabled
			multirefs := strings.Split(t.Field(i).Tag.Get("multiref"), ",")
			names = append(names, multirefs...)
		}

		for _, name := range names {
			tTF[name] = v.Field(i)
		}
	}

	return tTF
}

// AsSimple returns all information of a Info struct as a simple
// key/value string.
func (info Info) AsSimple() string {
	ret := []string{}

	for key, field := range info.tagToField(false) {
		if field.String() != "" {
			ret = append(ret, fmt.Sprintf("%s='%s'", key, field.String()))
		}
	}

	// Sort for deterministic output
	sort.Strings(ret)

	// Handle and sort properties separately, since they are
	// position-dependant.
	props := []string{}
	for key, valueL := range info.ConnectProps {
		if len(valueL) == 0 {
			props = append(props, key+"=''")
		} else {
			props = append(props, fmt.Sprintf("%s='%s'", key, valueL[len(valueL)-1]))
		}
	}

	sort.Strings(props)

	return strings.Join(append(ret, props...), " ")
}

// Prop returns the last value for a property or empty string.
// To access other values use ConnectProps directly.
func (info Info) Prop(property string) string {
	if info.ConnectProps == nil {
		return ""
	}

	vals, ok := info.ConnectProps[property]
	if !ok {
		return ""
	}

	if len(vals) == 0 {
		return ""
	}

	return vals[len(vals)-1]
}

// PropDefault calls .Prop with property and returns the result if it is
// not empty and defaultValue otherwise.
func (info Info) PropDefault(property, defaultValue string) string {
	if val := info.Prop(property); val != "" {
		return val
	}

	return defaultValue
}
