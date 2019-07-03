package libtest

import (
	"database/sql"

	"github.com/SAP/go-ase/libase/types"

	"testing"
)

// DoTestDecimal3838 tests the handling of the Decimal3838.
func DoTestDecimal3838(t *testing.T) {
	TestForEachDB("TestDecimal3838", t, testDecimal3838)
	//
}

func testDecimal3838(t *testing.T, db *sql.DB, tableName string) {
	pass := make([]interface{}, len(samplesDecimal3838))
	mySamples := make([]*types.Decimal, len(samplesDecimal3838))

	for i, sample := range samplesDecimal3838 {

		// Convert sample with passed function before proceeding
		mySample, err := types.NewDecimalString(sample)
		if err != nil {
			t.Errorf("Failed to convert sample %v: %v", sample, err)
			return
		}

		pass[i] = mySample
		mySamples[i] = mySample
	}

	rows, err := SetupTableInsert(db, tableName, "decimal(38,38)", pass...)
	if err != nil {
		t.Errorf("Error preparing table: %v", err)
		return
	}
	defer rows.Close()

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
