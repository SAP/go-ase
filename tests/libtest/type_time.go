package libtest

import (
	"database/sql"
	"testing"
	"time"
)

// DoTestTime tests that the time.Time type is handled correctly.
func DoTestTime(t *testing.T) {
	TestForEachDB("TestTime", t, testTime)
}

func testTime(t *testing.T, db *sql.DB, tableName string) {
	samples := []time.Time{
		// Sybase & Golang zero-value; January 1, 0001 Midnight
		time.Time{},
		time.Date(2019, time.March, 29, 9, 26, 0, 0, time.UTC),
		// Sybase max
		time.Date(9999, time.December, 31, 23, 59, 59, 999999000, time.UTC),
	}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}

	rows, err := SetupTableInsert(db, tableName, "bigdatetime", pass...)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv time.Time
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
