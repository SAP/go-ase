package libtest

import (
	"database/sql"

	"testing"
)

// DoTestUnsignedInt tests the handling of the UnsignedInt.
func DoTestUnsignedInt(t *testing.T) {
	TestForEachDB("TestUnsignedInt", t, testUnsignedInt)
	//
}

func testUnsignedInt(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesUnsignedInt))
	mySamples := make([]uint32, len(samplesUnsignedInt))

	for i, sample := range samplesUnsignedInt {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, teardownFn, err := SetupTableInsert(db, tableName, "unsigned int", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()
	defer teardownFn()

	i := 0
	var recv uint32
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
