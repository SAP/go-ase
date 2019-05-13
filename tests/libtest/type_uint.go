package libtest

import (
	"database/sql"
	"math"
	"testing"
)

// DoTestUint64 tests that the uint64 type is handled correctly.
func DoTestUint64(t *testing.T) {
	TestForEachDB("TestUint64", t, testUint64)
}

func testUint64(t *testing.T, db *sql.DB, tableName string) {
	samples := []uint64{0, 1000, 5000, 150000, 123456789, math.MaxUint32 + 1}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}
	rows, err := SetupTableInsert(db, tableName, "unsigned bigint", pass...)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv uint64
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

		i++
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
