package libtest

import (
	"database/sql"

	"testing"
)

// DoTestTinyInt tests the handling of the TinyInt.
func DoTestTinyInt(t *testing.T) {
	TestForEachDB("TestTinyInt", t, testTinyInt)
	//
}

func testTinyInt(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesTinyInt))
	mySamples := make([]uint8, len(samplesTinyInt))

	for i, sample := range samplesTinyInt {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "tinyint", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv uint8
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		if recv != mySamples[i] {

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
