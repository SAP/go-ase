package libtests

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func DBTestFunc(t *testing.T, db *sql.DB, tableName string)

func TestForEachDB(testName string, t *testing.T, testFn DBTestFunc) {
	dbs, err := libtest.GetDBs()
	if err != nil {
		t.Errorf("Error retrieving DBs: %v", err)
		return
	}

	for connectName, db := range dbs {
		defer db.Close()
		t.Run(connectName,
			func(t *testing.T) {
				testInt64(t, db, connectName+testName)
			},
		)
	}
}

func RandomNumber() string {
	return strconv.Itoa(rand.Int())
}

func LogDefer(fn func() error) {
	err := fn()
	if err != nil {
		log.Printf("Error in deferred function: %v", err)
	}
}
