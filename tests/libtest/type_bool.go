package libtest

import (
	"database/sql"
	"testing"
)

func DoTestBool(t *testing.T) {
	TestForEachDB("TestBool", t, testBool)
}

func testBool(t *testing.T, db *sql.DB, tableName string) {
	samples := []bool{true, false}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}

	rows, err := SetupTableInsert(db, tableName, "bit default 0 not null", pass...)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv bool
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
