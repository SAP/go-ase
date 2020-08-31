// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package libtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

// DoTestSQLExec runs tests for sql.Exec.
func DoTestSQLExec(t *testing.T) {
	t.Run("no cancel",
		func(t *testing.T) {
			TestForEachDB("TestSQLExecNoCancel", t, testSQLExecNoCancel)
		},
	)

	t.Run("cancel",
		func(t *testing.T) {
			TestForEachDB("TestSQLExecCancel", t, testSQLExecCancel)
		},
	)
}

func testSQLExecNoCancel(t *testing.T, db *sql.DB, tableName string) {
	_, err := db.ExecContext(context.Background(), fmt.Sprintf("create table %s (a int)", tableName))
	if err != nil {
		t.Errorf("Error creating table: %s: %v", tableName, err)
		return
	}

	_, err = db.ExecContext(context.Background(), fmt.Sprintf("insert into %s (a) values (?)", tableName), 5)
	if err != nil {
		t.Errorf("Error inserting value: %v", err)
		return
	}
}

func testSQLExecCancel(t *testing.T, db *sql.DB, tableName string) {
	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err := db.ExecContext(ctx, fmt.Sprintf("create table %s (a int)", tableName))
	if err != ctx.Err() {
		t.Errorf("Did not receive context error: %v", err)
	}
}
