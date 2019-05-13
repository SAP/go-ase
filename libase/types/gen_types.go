// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open("../../cgo/includes/cstypes.h")
	if err != nil {
		log.Printf("Failed to open cstypes.h: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Keep references from name to integer
	typeMap := map[string]int{}
	// Record all types in a string slice to sort after parsing all
	// types. This is used to generate reproducible maps, preventing
	// a different order with each run.
	typeSlice := sort.StringSlice{}

	for scanner.Scan() {
		str := scanner.Text()
		if !strings.HasPrefix(str, "#define") {
			continue
		}

		split := strings.Fields(str)

		// Verify the three parts '#define CS_x_TYPE (CS_INT)x'
		if len(split) != 3 {
			continue
		}

		// Verify second part 'CS_x_TYPE'
		if !strings.HasPrefix(split[1], "CS_") || !strings.HasSuffix(split[1], "_TYPE") {
			continue
		}

		// CS_<x>_TYPE
		key := strings.Split(split[1], "_")[1]
		// (CS_INT)<x> or (CS_INT)(<x>)
		value := strings.Split(split[2], ")")[1]
		// <x> or (<x>) -> <x>
		value = strings.Trim(value, "()")
		// Convert to integer
		valueI, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Failed to parse '%s' as integer: %v", value, err)
			return
		}

		typeMap[key] = valueI
		typeSlice = append(typeSlice, key)
	}
	typeSlice.Sort()

	// Write types.go
	buf := bytes.Buffer{}

	buf.WriteString("package types\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"reflect\"\n")
	buf.WriteString("\t\"time\"\n")
	buf.WriteString(")\n\n")

	// Write constants
	buf.WriteString("const (\n")
	for _, key := range typeSlice {
		val := typeMap[key]
		buf.WriteString(fmt.Sprintf("    %s ASEType = %d\n", key, val))
	}
	buf.WriteString(")\n\n")

	// Write type maps
	buf.WriteString("var string2type = map[string]ASEType{\n")
	for _, key := range typeSlice {
		buf.WriteString(fmt.Sprintf("    \"%s\": %s,\n", key, key))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var type2string = map[ASEType]string{\n")
	for _, key := range typeSlice {
		buf.WriteString(fmt.Sprintf("    %s: \"%s\",\n", key, key))
	}
	buf.WriteString("}\n\n")

	type2reflect := map[string]string{}
	type2interface := map[string]string{}

	for _, key := range typeSlice {
		switch key {
		// binary
		case "BINARY", "LONGBINARY", "VARBINARY":
			type2reflect[key] = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
			type2interface[key] = "[]byte{0}"
		// char, varchar
		// CTlib handles varchar as an array of char with the same type
		// CS_CHAR_TYPE, hence char types must also be handles as
		// string, otherwise only the first char of a string would be
		// returned.
		case "CHAR", "VARCHAR", "LONGCHAR", "UNICHAR":
			type2reflect[key] = "reflect.TypeOf(string(\"\"))"
			type2interface[key] = "string(\"\")"
		// bit
		case "BIT":
			type2reflect[key] = "reflect.TypeOf(uint64(0))"
			type2interface[key] = "uint64(0)"
		// xml
		case "XML":
			type2reflect[key] = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
			type2interface[key] = "[]byte{0}"
		// datetime
		case "DATE", "TIME", "DATETIME4":
			type2reflect[key] = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
			type2interface[key] = "[]byte{0}"
		case "DATETIME", "BIGDATETIME", "BIGTIME":
			type2reflect[key] = "reflect.TypeOf(time.Time{})"
			type2interface[key] = "time.Time{}"
		// numeric
		case "TINYINT", "SMALLINT", "INT", "BIGINT":
			type2reflect[key] = "reflect.TypeOf(int64(0))"
			type2interface[key] = "int64(0)"
		case "USMALLINT", "USHORT", "UINT", "UBIGINT", "NUMERIC", "LONG":
			type2reflect[key] = "reflect.TypeOf(uint64(0))"
			type2interface[key] = "uint64(0)"
		case "DECIMAL", "FLOAT":
			type2reflect[key] = "reflect.TypeOf(float64(0))"
			type2interface[key] = "float64(0)"
		// money
		case "MONEY", "MONEY4":
			type2reflect[key] = "reflect.TypeOf(uint64(0))"
			type2interface[key] = "uint64(0)"
		// text
		case "TEXT", "UNITEXT":
			type2reflect[key] = "reflect.TypeOf(string(\"\"))"
			type2interface[key] = "string(\"\")"
		// image
		case "IMAGE", "BLOB":
			type2reflect[key] = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
			type2interface[key] = "[]byte{0}"
		default:
			type2reflect[key] = "nil"
			type2interface[key] = "nil"
		}

	}
	buf.WriteString("var type2reflect = map[ASEType]reflect.Type{\n")
	for _, key := range typeSlice {
		buf.WriteString(fmt.Sprintf("    %s: %s,\n", key, type2reflect[key]))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var type2interface = map[ASEType]interface{}{\n")
	for _, key := range typeSlice {
		buf.WriteString(fmt.Sprintf("    %s: %s,\n", key, type2interface[key]))
	}
	buf.WriteString("}\n\n")

	// Format buffer
	formattedBuf, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Formatting code failed: %v", err)
		return
	}

	// Write result to types.go
	err = ioutil.WriteFile("types.go", formattedBuf, 0644)
	if err != nil {
		log.Printf("Writing generated code to types.go failed: %v", err)
		return
	}
}
