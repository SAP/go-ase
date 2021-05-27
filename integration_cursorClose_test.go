// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package ase

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/SAP/go-dblib/integration"
)

func TestCursorClose(t *testing.T) {

	t.Run("SingeCursor", func(t *testing.T) {
		integration.TestForEachDB("TestCursorCloseSingleCursor", t, func(t *testing.T, db *sql.DB, tableName string) {
			wrapper(t, db, tableName, singleCursor)
		})
	})

	t.Run("SingeCursorWithArgs", func(t *testing.T) {
		integration.TestForEachDB("TestCursorCloseSingleCursorWithArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			wrapper(t, db, tableName, singleCursorWithArgs)
		})
	})

	t.Run("TwoCursors", func(t *testing.T) {
		integration.TestForEachDB("TestCursorCloseTwoCursors", t, func(t *testing.T, db *sql.DB, tableName string) {
			wrapper(t, db, tableName, twoCursors)
		})
	})

	t.Run("TwoCursorsOneWithArgs", func(t *testing.T) {
		integration.TestForEachDB("TestCursorCloseTwoCursorsOneWithArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			wrapper(t, db, tableName, twoCursorsOneWithArgs)
		})
	})

	t.Run("TwoCursorsWithArgs", func(t *testing.T) {
		integration.TestForEachDB("TestCursorCloseTwoCursorsWithArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			wrapper(t, db, tableName, twoCursorsWithArgs)
		})
	})

	t.Run("TwoCursorsOneWithArgs2", func(t *testing.T) {
		integration.TestForEachDB("TestCursorCloseTwoCursorsOneWithArgs2", t, func(t *testing.T, db *sql.DB, tableName string) {
			wrapper(t, db, tableName, twoCursorsOneWithArgs2)
		})
	})

}

func createTable(db *sql.DB, tableName string) error {
	if _, err := db.Exec("create table " + tableName + " (a int, b varchar(30))"); err != nil {
		return fmt.Errorf("error creating table %s: %w", tableName, err)
	}

	for i, val := range []string{"one", "two", "three", "four"} {
		if _, err := db.Exec("insert into "+tableName+" (a, b) values (?, ?)", i+1, val); err != nil {
			return fmt.Errorf("error inserting values (%d, %s) into table %s: %w", i+1, val, tableName, err)
		}
	}

	return nil
}

func wrapper(t *testing.T, db *sql.DB, tableName string, runner func(*testing.T, *Conn, string)) {
	if err := createTable(db, tableName); err != nil {
		t.Errorf("error creating table: %v", err)
		return
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Errorf("error getting conn from sql.DB: %v", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("error closing conn from sql.DB: %v", err)
		}
	}()

	conn.Raw(func(driverConn interface{}) error {
		aseConn, ok := driverConn.(*Conn)
		if !ok {
			t.Errorf("received driverConn is not *Conn: %v", err)
			return nil
		}

		runner(t, aseConn, tableName)
		return nil
	})
}

func fetch(t *testing.T, cursor *Cursor) {
	rows, err := cursor.Fetch(context.Background())
	if err != nil {
		t.Errorf("error fetching result set: %v", err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Errorf("error closing rows: %v", err)
		}
	}()

	values := []driver.Value{0, ""}
	for {
		if err := rows.Next(values); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Errorf("error reading row: %v", err)
			return
		}

		fmt.Printf("| %d | %s |\n", values[0], values[1])
	}

	if err := rows.Close(); err != nil {
		t.Errorf("error closing rows: %v", err)
	}
}

func singleCursor(t *testing.T, conn *Conn, tableName string) {
	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName)
	if err != nil {
		t.Errorf("error creating cursor: %v", err)
		return
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			t.Errorf("error closing cursor: %v", err)
		}
	}()

	fetch(t, cursor)
}

func singleCursorWithArgs(t *testing.T, conn *Conn, tableName string) {
	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName+" where b like (?)", "two")
	if err != nil {
		t.Errorf("error creating cursor: %v", err)
		return
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			t.Errorf("error closing cursor: %v", err)
		}
	}()

	fetch(t, cursor)
}

func twoCursors(t *testing.T, conn *Conn, tableName string) {
	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName)
	if err != nil {
		t.Errorf("error creating cursor: %v", err)
		return
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			t.Errorf("error closing cursor: %v", err)
		}
	}()

	fetch(t, cursor)

	cursor2, err := conn.NewCursor(context.Background(), "select * from "+tableName)
	if err != nil {
		t.Errorf("error creating cursor2: %v", err)
		return
	}
	defer func() {
		if err := cursor2.Close(context.Background()); err != nil {
			t.Errorf("error closing cursor2: %v", err)
		}
	}()

	fetch(t, cursor2)
}

func twoCursorsOneWithArgs(t *testing.T, conn *Conn, tableName string) {
	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName)
	if err != nil {
		t.Errorf("error creating cursor: %v", err)
		return
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			t.Errorf("error closing cursor: %v", err)
		}
	}()

	fetch(t, cursor)

	cursorWithArgs, err := conn.NewCursor(context.Background(), "select * from "+tableName+" where b like (?)", "two")
	if err != nil {
		t.Errorf("error creating cursorWithArgs: %v", err)
		return
	}
	defer func() {
		if err := cursorWithArgs.Close(context.Background()); err != nil {
			t.Errorf("error closing cursorWithArgs: %v", err)
		}
	}()

	fetch(t, cursorWithArgs)
}

func twoCursorsOneWithArgs2(t *testing.T, conn *Conn, tableName string) {
	cursorWithArgs, err := conn.NewCursor(context.Background(), "select * from "+tableName+" where b like (?)", "two")
	if err != nil {
		t.Errorf("error creating cursorWithArgs: %v", err)
		return
	}
	defer func() {
		if err := cursorWithArgs.Close(context.Background()); err != nil {
			t.Errorf("error closing cursorWithArgs: %v", err)
		}
	}()

	fetch(t, cursorWithArgs)

	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName)
	if err != nil {
		t.Errorf("error creating cursor: %v", err)
		return
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			t.Errorf("error closing cursor: %v", err)
		}
	}()

	fetch(t, cursor)
}

func twoCursorsWithArgs(t *testing.T, conn *Conn, tableName string) {
	cursorWithArgs, err := conn.NewCursor(context.Background(), "select * from "+tableName+" where b like (?)", "two")
	if err != nil {
		t.Errorf("error creating cursorWithArgs: %v", err)
		return
	}
	defer func() {
		if err := cursorWithArgs.Close(context.Background()); err != nil {
			t.Errorf("error closing cursorWithArgs: %v", err)
		}
	}()

	fetch(t, cursorWithArgs)

	cursorWithArgs2, err := conn.NewCursor(context.Background(), "select * from "+tableName+" where b like (?)", "two")
	if err != nil {
		t.Errorf("error creating cursorWithArgs2: %v", err)
		return
	}
	defer func() {
		if err := cursorWithArgs2.Close(context.Background()); err != nil {
			t.Errorf("error closing cursorWithArgs2: %v", err)
		}
	}()

	fetch(t, cursorWithArgs2)
}
