package libtest

import (
	"database/sql"
	"math"
	"testing"
)

// DoTestFloat64 tests that the float64 type is handled correctly.
func DoTestFloat64(t *testing.T) {
	TestForEachDB("TestFloat64", t, testFloat64)
}

func testFloat64(t *testing.T, db *sql.DB, tableName string) {
	samples := []float64{math.SmallestNonzeroFloat64, math.MaxFloat64}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}
	rows, err := SetupTableInsert(db, tableName, "float", pass...)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv float64
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
