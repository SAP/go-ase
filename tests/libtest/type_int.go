package libtest

import (
	"database/sql"
	"fmt"
	"math"
	"testing"
)

func DoTestInt64(t *testing.T) {
	TestForEachDB("TestInt64", t, testInt64)
}

func testInt64(t *testing.T, db *sql.DB, tableName string) {

	_, err := db.Exec("create table ? (a bigint)", tableName)
	if err != nil {
		t.Errorf("Failed to create table %s: %v", tableName, err)
		return
	}

	stmt, err := db.Prepare(fmt.Sprintf("insert into %s values (?)", tableName))
	if err != nil {
		t.Errorf("Failed to prepare statement: %v", err)
		return
	}
	defer stmt.Close()

	samples := []int64{math.MinInt64, -5000, -100, 0, 100, 5000, math.MaxInt64}
	for _, sample := range samples {
		_, err := stmt.Exec(sample)
		if err != nil {
			t.Errorf("Failed to execute prepared statement with %v: %v", sample, err)
			return
		}
	}

	rows, err := db.Query("select * from ?", tableName)
	if err != nil {
		t.Errorf("Error selecting from TestInt64: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv int64
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		if recv != samples[i] {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", samples[i])
			t.Errorf("Received: %v", recv)
		}

		i += 1
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
