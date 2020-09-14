// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/SAP/go-ase"
	"github.com/SAP/go-dblib/dsn"
)

var (
	dbName    = "testCursor"
	tableName = "test"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("cursor: %v", err)
	}
}

func DoMain() error {
	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	fmt.Println("preparing table")
	if _, err = db.Exec(fmt.Sprintf("if object_id('%s') is not null drop table %s", tableName, tableName)); err != nil {
		return fmt.Errorf("failed to drop table %q: %w", tableName, err)
	}

	if _, err = db.Exec(fmt.Sprintf("create table %s (a int, b varchar(30))", tableName)); err != nil {
		return fmt.Errorf("failed to create table %q: %w", tableName, err)
	}

	if _, err = db.Exec("insert into "+tableName+" values (?, ?)", 1, "one"); err != nil {
		return fmt.Errorf("failed to insert values into table %q: %w", tableName, err)
	}

	if _, err = db.Exec("insert into "+tableName+" values (?, ?)", 2, "two"); err != nil {
		return fmt.Errorf("failed to insert values into table %q: %w", tableName, err)
	}

	fmt.Println("inserted values:")
	rows, err := db.Query("select * from " + tableName)
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to retrieve column names: %w", err)
	}

	fmt.Printf("| %-10s | %-30s |\n", colNames[0], colNames[1])
	format := "| %-10d | %-30s |\n"

	var a int
	var b string

	for rows.Next() {
		if err = rows.Scan(&a, &b); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Printf(format, a, b)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error getting conn: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing conn: %v", err)
		}
	}()

	return conn.Raw(func(driverConn interface{}) error {
		if err := rawProcess(driverConn); err != nil {
			return fmt.Errorf("error in rawProcess: %w", err)
		}
		return nil
	})
}

func rawProcess(driverConn interface{}) error {
	conn, ok := driverConn.(*ase.Conn)
	if !ok {
		return errors.New("invalid driver, conn is not *github.com/SAP/go-ase.Conn")
	}

	fmt.Println("opening cursor1")
	cursor, err := conn.NewCursor(context.Background(), "select * from test")
	if err != nil {
		return fmt.Errorf("error creating cursor: %w", err)
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Printf("error closing cursor: %v", err)
		}
	}()

	fmt.Println("fetching cursor1")
	rows, err := cursor.Fetch(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching rows: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	loop := 0

	fmt.Println("iterating over cursor1")
	values := []driver.Value{0, ""}
	for {
		if err := rows.Next(values); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading row: %w", err)
		}

		fmt.Printf("a: %d\n", values[0])
		fmt.Printf("b: %s\n", values[1])

		loop++
		if loop > 5 {
			return nil
		}
	}

	fmt.Println("opening cursor2")
	cursor2, err := conn.NewCursor(context.Background(),
		"select a, b from test where b like (?)", "two")
	if err != nil {
		return fmt.Errorf("error creating cursor: %w", err)
	}
	defer func() {
		if err := cursor2.Close(context.Background()); err != nil {
			log.Printf("error closing cursor: %v", err)
		}
	}()

	fmt.Println("fetching cursor2")
	rows2, err := cursor2.Fetch(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching rows: %w", err)
	}
	defer func() {
		if err := rows2.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	loop = 0

	fmt.Println("iterating over cursor2")
	values = []driver.Value{0, ""}
	for {
		if err := rows2.Next(values); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading row: %w", err)
		}

		fmt.Printf("a: %d\n", values[0])
		fmt.Printf("b: %s\n", values[1])

		loop++
		if loop > 5 {
			return nil
		}
	}

	return nil
}
