package cgo

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestUint32(t *testing.T) {
	params := []uint32{0, 100, 1000, 25000}

	passParams := make([]interface{}, len(params))
	for i, param := range params {
		passParams[i] = param
	}
	recvParams := make([]interface{}, len(params))

	validateFn := func(i int) {
		val, ok := recvParams[i].(uint64)
		if !ok {
			t.Errorf("Could not convert cell value %#v to uint32, should be %v", recvParams[i], params[i])
			return
		}
		valC := uint32(val)

		if valC != params[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", params[i])
			t.Errorf("Received: %v", valC)
		}
	}

	testHelper(t, "unsigned int", "Uint32", passParams, recvParams, validateFn)
}

func TestUint64(t *testing.T) {
	params := []uint64{0, 100, 1000, 25000, math.MaxUint32 + 1}

	passParams := make([]interface{}, len(params))
	for i, param := range params {
		passParams[i] = param
	}
	recvParams := make([]interface{}, len(params))

	validateFn := func(i int) {
		val, ok := recvParams[i].(uint64)
		if !ok {
			t.Errorf("Could not convert cell value %#v to uint64, should be %v", recvParams[i], params[i])
			return
		}

		if val != params[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", params[i])
			t.Errorf("Received: %v", val)
		}
	}

	testHelper(t, "unsigned bigint", "Uint64", passParams, recvParams, validateFn)
}

func TestInt32(t *testing.T) {
	params := []int32{math.MinInt32, -5000, -100, 0, 100, 5000, math.MaxInt32}
	passParams := make([]interface{}, len(params))
	for i, param := range params {
		passParams[i] = param
	}
	recvParams := make([]interface{}, len(params))

	validateFn := func(i int) {
		val, ok := recvParams[i].(int64)
		if !ok {
			t.Errorf("Could not convert cell value %#v to uint32, should be %v", recvParams[i], params[i])
			return
		}
		valC := int32(val)

		if valC != params[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", params[i])
			t.Errorf("Received: %v", valC)
		}
	}

	testHelper(t, "int", "Int32", passParams, recvParams, validateFn)
}

func TestInt64(t *testing.T) {
	params := []int64{math.MinInt64, -5000, -100, 0, 100, 5000, math.MaxInt64}
	passParams := make([]interface{}, len(params))
	for i, param := range params {
		passParams[i] = param
	}
	recvParams := make([]interface{}, len(params))

	validateFn := func(i int) {
		if recvParams[i] != params[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", params[i])
			t.Errorf("Received: %v", recvParams[i])
		}
	}

	testHelper(t, "bigint", "Int64", passParams, recvParams, validateFn)
}

func TestInt(t *testing.T) {
	params := []int{math.MinInt64, -5000, -100, 0, 100, 5000, math.MaxInt64}
	passParams := make([]interface{}, len(params))
	for i, param := range params {
		passParams[i] = param
	}
	recvParams := make([]interface{}, len(params))

	validateFn := func(i int) {
		val := int(recvParams[i].(int64))

		if val != params[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", params[i])
			t.Errorf("Received: %v", val)
		}
	}

	testHelper(t, "bigint", "Int", passParams, recvParams, validateFn)
}

func TestString(t *testing.T) {
	// 255 is C.CS_MAX_CHAR but can't be referenced as such since go
	// doesn't like using cgo in tests.
	stringLengths := []int{0, 50, 150, 255}
	params := []string{}

	for _, length := range stringLengths {
		chrs := make([]rune, length)
		chr := 'a'
		for i := 0; i < length; i++ {
			chrs[i] = chr
			chr++
			if chr > 'z' {
				chr -= 26
			}
		}
		params = append(params, string(chrs))
	}

	passParams := make([]interface{}, len(params))
	for i, param := range params {
		passParams[i] = param
	}
	recvParams := make([]interface{}, len(params))

	validateFn := func(i int) {
		val := strings.TrimSpace(recvParams[i].(string))
		if val != params[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %#v", params[i])
			t.Errorf("Received: %#v", val)
		}
	}

	testHelper(t, "char(255)", "String", passParams, recvParams, validateFn)
}

func testHelper(t *testing.T,
	aseType string, testName string,
	params []interface{},
	recvParams []interface{},
	validateCell func(int),
) {
	db, err := sql.Open("ase", testDsn.AsSimple())
	if err != nil {
		t.Errorf("Error opening connection: %v", err)
		return
	}
	defer db.Close()

	// Create a table with on column of the type to be tested
	tableName := fmt.Sprintf("Test%s", testName)
	_, err = db.Exec("create table ? (a ?)", tableName, aseType)
	if err != nil {
		t.Errorf("Failed to create table: %v", err)
		return
	}

	// Insert all values to test through a prepared statement
	stmt, err := db.Prepare(fmt.Sprintf("insert into %s values (?)", tableName))
	if err != nil {
		t.Errorf("Failed to prepare statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, param := range params {
		_, err := stmt.Exec(param)
		if err != nil {
			t.Errorf("Error when executing prepared statement with %v: %v", param, err)
			return
		}
	}

	// Read values back out
	rows, err := db.Query("select * from ?", tableName)
	if err != nil {
		t.Errorf("Error selecting from %s: %v", tableName, err)
		return
	}
	defer rows.Close()

	// Retrieve rows and validate each cell
	i := 0
	for rows.Next() {
		err = rows.Scan(&recvParams[i])
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			return
		}

		validateCell(i)
		i += 1
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
