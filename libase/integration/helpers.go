// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"database/sql"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// DBTestFunc is the interface for tests accepting a pre-connected
// sql.DB.
type DBTestFunc func(t *testing.T, db *sql.DB, tableName string)

// TestForEachDB runs the given DBTestFunc against all registered
// connection types.
func TestForEachDB(testName string, t *testing.T, testFn DBTestFunc) {
	for connectName, dbFn := range sqlDBMap {
		db, err := dbFn()
		if err != nil {
			t.Errorf("Connection failed for '%s': %v", connectName, err)
			continue
		}
		defer db.Close()

		t.Run(connectName,
			func(t *testing.T) {
				testFn(t, db, strings.Replace(testName+connectName, " ", "_", -1))
			},
		)
	}
}

// RandomNumber returns an unsecure random number as a string.
//
// This method is used to ensure random names for similar objects being
// created for testing purposes in databases.
func RandomNumber() string {
	return strconv.Itoa(rand.Int())
}
