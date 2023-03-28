// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strconv"
	"sync"
	"testing"

	"github.com/SAP/go-dblib/dsn"
)

var (
	tablePrepared = &sync.Once{}
	tablename     = "benchmark"
	query         = "select a, b from " + tablename
)

func prepare(b *testing.B, fn func(*testing.B, *Conn)) {
	info, err := NewInfoWithEnv()
	if err != nil {
		b.Errorf("error reading DSN info from env: %v", err)
		return
	}

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		b.Errorf("failed to open connection to database: %v", err)
		return
	}
	defer db.Close()

	errored := false
	tablePrepared.Do(func() {
		if _, err = db.Exec("if object_id('" + tablename + "') is not null drop table " + tablename); err != nil {
			b.Errorf("error dropping table if exists: %v", err)
			errored = true
			return
		}

		if _, err := db.Exec("create table " + tablename + " (a bigint, b varchar(100))"); err != nil {
			b.Errorf("error creating table: %v", err)
			errored = true
			return
		}

		stmt, err := db.Prepare("insert into " + tablename + " values (?, ?)")
		if err != nil {
			b.Errorf("error preparing insert statement: %v", err)
			errored = true
			return
		}
		defer stmt.Close()

		for i := 0; i < 1000000; i++ {
			iStr := strconv.Itoa(i)
			if _, err := stmt.Exec(i, iStr); err != nil {
				b.Errorf("error inserting values (%d, %s): %v", i, iStr, err)
				errored = true
				return
			}
		}
	})
	if errored {
		return
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		b.Errorf("error getting conn: %v", err)
		return
	}
	defer conn.Close()

	_ = conn.Raw(func(driverConn interface{}) error {
		conn, ok := driverConn.(*Conn)
		if !ok {
			b.Errorf("driverConn is not *Conn, is %T", driverConn)
			return nil
		}

		fn(b, conn)
		return nil
	})
}

func BenchmarkCursorRows_Next(b *testing.B) {
	prepare(b, cursorRows_Next)
}

func cursorRows_Next(b *testing.B, conn *Conn) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cursor, err := conn.NewCursor(context.Background(), query)
		if err != nil {
			b.Errorf("error creating cursor: %v", err)
			return
		}
		defer cursor.Close(context.Background())

		rows, err := cursor.Fetch(context.Background())
		if err != nil {
			b.Errorf("error fetching rows: %v", err)
			return
		}
		defer rows.Close()

		values := []driver.Value{0, ""}
		for {
			if err := rows.Next(values); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				b.Errorf("error scanning fields: %v", err)
				return
			}
		}
	}
}

func BenchmarkRows_Next(b *testing.B) {
	prepare(b, rows_Next)
}

func rows_Next(b *testing.B, conn *Conn) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, _, err := conn.DirectExec(context.Background(), query)
		if err != nil {
			b.Errorf("error executing statement: %v", err)
			return
		}

		values := []driver.Value{0, ""}
		for {
			if err := rows.Next(values); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				b.Errorf("error scanning fields: %v", err)
				return
			}
		}
	}
}
