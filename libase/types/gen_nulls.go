// +build ignore

package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

var (
	fType   = flag.String("type", "", "Type to generate a NullType for")
	fGoType = flag.String("gotype", "", "Go equivalent for the type")
	fImport = flag.String("import", "", "Package to import")
)

var sTemplate = `package types

import (
	"database/sql"
	"database/sql/driver"
	{{ if .Import }}"{{.Import}}"{{ end }}
)

var (
	_ driver.Valuer = (*Null{{.T}})(nil)
	_ sql.Scanner = (*Null{{.T}})(nil)
)

type Null{{.T}} struct {
	{{.T}} {{.GoT}}
	Valid bool
}

func (null{{.T}} *Null{{.T}}) Scan(value interface{}) error {
	if value == nil {
		null{{.T}}.{{.T}} = {{.GoT}}{}
		null{{.T}}.Valid = false
		return nil
	}


	null{{.T}}.{{.T}}, null{{.T}}.Valid = value.({{.GoT}})
	return nil
}

func (null{{.T}} Null{{.T}}) Value() (driver.Value, error) {
	if !null{{.T}}.Valid {
		return nil, nil
	}

	return null{{.T}}.{{.T}}, nil
}
`

type data struct {
	T, GoT, Import string
}

func main() {
	flag.Parse()
	if *fType == "" || *fGoType == "" {
		log.Printf("Both -type and -gotype are required")
		os.Exit(1)
	}

	d := data{
		T:      *fType,
		GoT:    *fGoType,
		Import: *fImport,
	}

	tmpl, err := template.New("").Parse(sTemplate)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		os.Exit(1)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, d)
	if err != nil {
		log.Printf("Failed to execute template with data '%v': %v", d, err)
		os.Exit(1)
	}

	formattedBuf, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Formatting code failed: %v", err)
		os.Exit(1)
	}

	outfileName := "null" + *fType + ".go"
	err = ioutil.WriteFile(outfileName, formattedBuf, 0644)
	if err != nil {
		log.Printf("Failed to write nulltype definition to '%s': %v", outfileName, err)
		os.Exit(1)
	}
}
