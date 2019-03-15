package dsn

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

// DsnInfo represents all required information to open a connection to
// an ASE server.
//
// The json tag is the expected string in a simple URI.
type DsnInfo struct {
	Host         string     `json:"host,hostname" validate:"required"`
	Port         string     `json:"port" validate:"required"`
	Username     string     `json:"user,username" validate:"required"`
	Password     string     `json:"pass,passwd,password" validate:"required"`
	Userstorekey string     `json:"userstorekey" validate:"required"`
	Database     string     `json:"db,database"`
	ConnectProps url.Values `json:"connectprops"`
}

// tagToField returns a mapping from json metadata tags to
// reflect.Values.
// If multiref is true each json metadata tag will be mapped to its
// field.Value; If multref is false only the first json metadata tag
// will be mapped:
// multiref = true:
//   map[string]reflect.Value{"host": dsnInfo.Host, "hostname": dsnInfo.Host}
// multiref = false:
//   map[string]reflect.Value{"host": dsnInfo.Host}
func (dsnInfo *DsnInfo) tagToField(multiref bool) map[string]reflect.Value {
	tTF := map[string]reflect.Value{}
	// The accepted type of ValueOf is interface, which still allows
	// accessing the metadata but not the fields, since an interface
	// doesn't have field.
	// By passing a pointer it is possible to call .Elem(), which
	// returns a reflect.Value representation of the passed struct
	// - which allows to access its fields.
	v := reflect.ValueOf(dsnInfo).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		// Skip over ConnectProps, this member is handled specially
		// since it is a map with additional methods.
		if t.Field(i).Name == "ConnectProps" {
			continue
		}

		// Allow attributes such as json:"hostname,host" to map
		// "hostname" and "host" to `.Hostname`.
		names := strings.Split(t.Field(i).Tag.Get("json"), ",")
		if multiref {
			for _, name := range names {
				tTF[name] = v.Field(i)
			}
		} else {
			tTF[names[0]] = v.Field(i)
		}
	}

	return tTF
}

func (dsnInfo DsnInfo) AsSimple() string {
	ret := []string{}

	for key, field := range dsnInfo.tagToField(false) {
		if field.String() != "" {
			ret = append(ret, fmt.Sprintf("%s='%s'", key, field.String()))
		}
	}

	for key, valueL := range dsnInfo.ConnectProps {
		for _, value := range valueL {
			ret = append(ret, fmt.Sprintf("%s='%s'", key, value))
		}
	}

	sort.Strings(ret)

	return strings.Join(ret, " ")
}
