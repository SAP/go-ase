package libtest

import (
	"database/sql"

	"testing"

	"time"
)

// DoTestSmallDateTime tests the handling of the SmallDateTime.
func DoTestSmallDateTime(t *testing.T) {
	TestForEachDB("TestSmallDateTime", t, testSmallDateTime)
	//
}

func testSmallDateTime(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesSmallDateTime))
	mySamples := make([]time.Time, len(samplesSmallDateTime))

	for i, sample := range samplesSmallDateTime {

		mySample := sample

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, teardownFn, err := SetupTableInsert(db, tableName, "smalldatetime", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()
	defer teardownFn()

	i := 0
	var recv time.Time
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
