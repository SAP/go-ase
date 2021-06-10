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
)

// wrapper is used to wrap tests for the underlying driver connection in
// integration tests.
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

// interface to match both Rows and CursorRows.
type sqlRows interface {
	Next([]driver.Value) error
}

// fetchRows expects the passed rows to return {int, string} and prints
// all rows to stdout.
//
// rows is not closed automatically.
func fetchRows(t *testing.T, rows sqlRows) {
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
}
