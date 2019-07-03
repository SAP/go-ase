package libtest

import (
	"database/sql"

	"testing"
)

// DoTestUnsignedSmallInt tests the handling of the UnsignedSmallInt.
func DoTestUnsignedSmallInt(t *testing.T) {
	TestForEachDB("TestUnsignedSmallInt", t, testUnsignedSmallInt)
	//
}

func testUnsignedSmallInt(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesUnsignedSmallInt))
	mySamples := make([]uint16, len(samplesUnsignedSmallInt))

	for i, sample := range samplesUnsignedSmallInt {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "unsigned smallint", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv uint16
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
