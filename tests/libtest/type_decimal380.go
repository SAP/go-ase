package libtest

import (
	"database/sql"

	"github.com/SAP/go-ase/libase/types"

	"testing"
)

// DoTestDecimal380 tests the handling of the Decimal380.
func DoTestDecimal380(t *testing.T) {
	TestForEachDB("TestDecimal380", t, testDecimal380)
	//
}

func testDecimal380(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesDecimal380))
	mySamples := make([]*types.Decimal, len(samplesDecimal380))

	for i, sample := range samplesDecimal380 {

		// Convert sample with passed function before proceeding
		mySample, err := convertDecimal380(sample)
		if err != nil {
			t.Errorf("Failed to convert sample %v: %v", sample, err)
			return
		}

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, teardownFn, err := SetupTableInsert(db, tableName, "decimal(38,0)", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()
	defer teardownFn()

	i := 0
	var recv *types.Decimal
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		if compareDecimal(recv, mySamples[i]) {

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
