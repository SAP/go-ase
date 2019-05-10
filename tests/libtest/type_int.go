package libtest

import (
	"database/sql"
	"math"
	"testing"
)

func DoTestInt64(t *testing.T) {
	TestForEachDB("TestInt64", t, testInt64)
}

func testInt64(t *testing.T, db *sql.DB, tableName string) {
	samples := []int64{math.MinInt64, -5000, -100, 0, 100, 5000, math.MaxInt64}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}
	rows, err := SetupTableInsert(db, tableName, "bigint", pass...)
	if err != nil {
		t.Errorf("%v", err)
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

		i++
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
