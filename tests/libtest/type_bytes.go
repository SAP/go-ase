package libtest

import (
	"bytes"
	"database/sql"
	"testing"
)

func DoTestBytes(t *testing.T) {
	TestForEachDB("TestBytes", t, testBytes)
}

func testBytes(t *testing.T, db *sql.DB, tableName string) {
	samples := [][]byte{
		[]byte{},
		[]byte("test"),
		[]byte("a longer test"),
	}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}

	rows, err := SetupTableInsert(db, tableName, "binary(255) null", pass...)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv []byte
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		recv = bytes.Trim(recv, "\x00")

		if bytes.Compare(recv, samples[i]) != 0 {
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
