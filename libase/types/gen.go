// +build ignore

package main

import (
	"bufio"
	"fmt"
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

		// Skip over illegal type '(CS_INT)(-1)'
		if strings.HasSuffix(split[2], ")") {
			continue
		}

		// CS_<x>_TYPE
		key := strings.Split(split[1], "_")[1]
		// (CS_INT)<x>
		value := strings.Split(split[2], ")")[1]
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
	file, err = os.OpenFile("./types.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Failed to open types.go for writing: %v", err)
		return
	}
	defer file.Close()

	file.WriteString("package types\n\n")
	file.WriteString("import (\n")
	file.WriteString("\t\"reflect\"\n")
	file.WriteString("\t\"time\"\n")
	file.WriteString(")\n\n")

	// Write constants
	file.WriteString("const (\n")
	for _, key := range typeSlice {
		val := typeMap[key]
		file.WriteString(fmt.Sprintf("    %s ASEType = %d\n", key, val))
	}
	file.WriteString(")\n\n")

	// Write type maps
	file.WriteString("var string2type = map[string]ASEType{\n")
	for _, key := range typeSlice {
		file.WriteString(fmt.Sprintf("    \"%s\": %s,\n", key, key))
	}
	file.WriteString("}\n\n")

	file.WriteString("var type2string = map[ASEType]string{\n")
	for _, key := range typeSlice {
		file.WriteString(fmt.Sprintf("    %s: \"%s\",\n", key, key))
	}
	file.WriteString("}\n\n")

	file.WriteString("var type2reflect = map[ASEType]reflect.Type{\n")
	for _, key := range typeSlice {
		value := "nil"
		switch key {
		// binary
		case "BINARY", "LONGBINARY", "VARBINARY":
			value = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
		// char
		case "CHAR":
			value = "reflect.TypeOf(rune(' '))"
		case "VARCHAR", "LONGCHAR", "UNICHAR":
			value = "reflect.TypeOf(string(\"\"))"
		// bit
		case "BIT":
			value = "reflect.TypeOf(uint64(0))"
		// xml
		case "XML":
			value = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
		// datetime
		case "DATE", "TIME", "DATETIME4":
			value = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
		case "DATETIME", "BIGDATETIME", "BIGTIME":
			value = "reflect.TypeOf(time.Time{})"
		// numeric
		case "TINYINT", "SMALLINT", "INT", "BIGINT":
			value = "reflect.TypeOf(int64(0))"
		case "USMALLINT", "USHORT", "UINT", "UBIGINT", "NUMERIC", "LONG":
			value = "reflect.TypeOf(uint64(0))"
		case "DECIMAL", "FLOAT":
			value = "reflect.TypeOf(float64(0))"
		// money
		case "MONEY", "MONEY4":
			value = "reflect.TypeOf(uint64(0))"
		// text
		case "TEXT", "UNITEXT":
			value = "reflect.TypeOf(string(\"\"))"
		// image
		case "IMAGE", "BLOB":
			value = "reflect.SliceOf(reflect.TypeOf(byte(0)))"
		}

		file.WriteString(fmt.Sprintf("    %s: %s,\n", key, value))
	}
	file.WriteString("}\n\n")
}
