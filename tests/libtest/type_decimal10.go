package libtest

import (
	"database/sql"

	"github.com/SAP/go-ase/libase/types"

	"testing"
)

// DoTestDecimal10 tests the handling of the Decimal10.
func DoTestDecimal10(t *testing.T) {
	TestForEachDB("TestDecimal10", t, testDecimal10)
	//
}

func testDecimal10(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesDecimal10))
	mySamples := make([]*types.Decimal, len(samplesDecimal10))

	for i, sample := range samplesDecimal10 {

		// Convert sample with passed function before proceeding
		mySample, err := types.NewDecimalString(sample)
		if err != nil {
			t.Errorf("Failed to convert sample %v: %v", sample, err)
			return
		}

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, teardownFn, err := SetupTableInsert(db, tableName, "decimal(1,0)", pass...)
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
