package libtest

import (
	"context"
	"database/sql"
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
	_, err := db.ExecContext(context.Background(), "create table ? (a int)", tableName)
	if err != nil {
		t.Errorf("Error creating table: %s: %v", tableName, err)
		return
	}

	_, err = db.ExecContext(context.Background(), "insert into ? values (5)", tableName)
	if err != nil {
		t.Errorf("Error inserting value: %v", err)
		return
	}
}

func testSQLExecCancel(t *testing.T, db *sql.DB, tableName string) {
	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err := db.ExecContext(ctx, "create table ? (a int)", tableName)
	if err != ctx.Err() {
		t.Errorf("Did not receive context error: %v", err)
	}
}
