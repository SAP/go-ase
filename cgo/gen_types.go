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
	file, err := os.Open("./includes/cstypes.h")
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

		// TODO is skipping ILLEGAL ok?
		if key == "ILLEGAL" {
			continue
		}

		typeSlice = append(typeSlice, key)
	}
	typeSlice.Sort()

	// Write types.go
	buf := bytes.Buffer{}

	buf.WriteString("package cgo\n\n")

	// Write constants
	buf.WriteString("const (\n")
	for _, key := range typeSlice {
		val := typeMap[key]
		buf.WriteString(fmt.Sprintf("    %s ASEType = %d\n", key, val))
	}
	buf.WriteString(")\n\n")

	// Write type maps
	buf.WriteString("var type2string = map[ASEType]string{\n")
	for _, key := range typeSlice {
		buf.WriteString(fmt.Sprintf("    %s: \"%s\",\n", key, key))
	}
	buf.WriteString("}\n\n")

	// Format buffer
	formattedBuf, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Formatting code failed: %v", err)
		os.Exit(1)
	}

	// Write result to types.go
	err = ioutil.WriteFile("typeConsts.go", formattedBuf, 0644)
	if err != nil {
		log.Printf("Writing generated code to types.go failed: %v", err)
		os.Exit(1)
	}
}
