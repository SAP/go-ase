package libtest

import (
	"database/sql"
	"math/rand"
	"strconv"
	"testing"
)

type DBTestFunc func(t *testing.T, db *sql.DB, tableName string)

func TestForEachDB(testName string, t *testing.T, testFn DBTestFunc) {
	dbs, err := GetDBs()
	if err != nil {
		t.Errorf("Error retrieving DBs: %v", err)
		return
	}

	for connectName, db := range dbs {
		defer db.Close()
		t.Run(connectName,
			func(t *testing.T) {
				testFn(t, db, testName)
			},
		)
	}
}

func RandomNumber() string {
	return strconv.Itoa(rand.Int())
}
