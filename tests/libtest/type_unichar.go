package libtest

import (
	"database/sql"

	"testing"
)

// DoTestUniChar tests the handling of the UniChar.
func DoTestUniChar(t *testing.T) {
	TestForEachDB("TestUniChar", t, testUniChar)
	//
}

func testUniChar(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesUniChar))
	mySamples := make([]string, len(samplesUniChar))

	for i, sample := range samplesUniChar {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, teardownFn, err := SetupTableInsert(db, tableName, "unichar(30) null", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()
	defer teardownFn()

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
