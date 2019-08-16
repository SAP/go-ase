package libtest

import (
	"database/sql"

	"testing"
)

// DoTestChar tests the handling of the Char.
func DoTestChar(t *testing.T) {
	TestForEachDB("TestChar", t, testChar)
	//
}

func testChar(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesChar))
	mySamples := make([]string, len(samplesChar))

	for i, sample := range samplesChar {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "char(13)", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv string
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		if compareChar(recv, mySamples[i]) {

			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", mySamples[i])
			t.Errorf("Received: %v", recv)
		}

		i++
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
