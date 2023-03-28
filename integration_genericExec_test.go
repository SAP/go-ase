// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package ase

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/SAP/go-dblib/integration"
)

func TestDirectExec(t *testing.T) {

	t.Run("SelectNoArgs", func(t *testing.T) {
		integration.TestForEachDB("TestDirectExecSelectNoArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			directExecWrapper(t, db, tableName, "select * from %s", nil)
		})
	})

	t.Run("SelectWithArgs", func(t *testing.T) {
		integration.TestForEachDB("TestDirectExecSelectWithArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			directExecWrapper(t, db, tableName, "select * from %s where b like ?", "three")
		})
	})

	t.Run("UpdateNoArgs", func(t *testing.T) {
		integration.TestForEachDB("TestDirectExecUpdateNoArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			directExecWrapper(t, db, tableName, "update %s set b = \"five\"")
		})
	})

	t.Run("UpdateWithArgs", func(t *testing.T) {
		integration.TestForEachDB("TestDirectExecUpdateWithArgs", t, func(t *testing.T, db *sql.DB, tableName string) {
			directExecWrapper(t, db, tableName, "update %s set b = \"five\" where b like ?", "three")
		})
	})

}

func directExecWrapper(t *testing.T, db *sql.DB, tableName, query string, args ...interface{}) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wrapper(t, db, tableName,
		func(t *testing.T, conn *Conn, tableName string) {
			rows, result, err := conn.DirectExec(timeout, fmt.Sprintf(query, tableName), args...)
			if err != nil {
				t.Errorf("received error on DirectExec: %v", err)
				return
			}

			if rows == nil {
				t.Errorf("received nil rows")
			} else {
				fetchRows(t, rows)
				if err := rows.Close(); err != nil {
					t.Errorf("error closing rows: %v", err)
				}
			}

			if result == nil {
				t.Errorf("received nil result")
			}
		},
	)
}
