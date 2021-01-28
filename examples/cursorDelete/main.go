// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
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
	tableName = "testCursorDelete"
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

	if err := prepare(db); err != nil {
		return err
	}

	defer func() {
		if err := teardown(db); err != nil {
			log.Printf("error during teardown: %v", err)
		}
	}()

	fmt.Println("before delete:")
	if err := printTable(db); err != nil {
		return fmt.Errorf("error printing table: %w", err)
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error getting conn from sql.DB: %w", err)
	}
	defer conn.Close()

	if err := conn.Raw(example); err != nil {
		return fmt.Errorf("error executing on raw connection: %w", err)
	}

	fmt.Println("after delete:")
	if err := printTable(db); err != nil {
		return fmt.Errorf("error printing table: %w", err)
	}

	return nil
}

// prepare creates a table and inserts example values.
func prepare(db *sql.DB) error {
	if err := teardown(db); err != nil {
		return err
	}

	if _, err := db.Exec("create table " + tableName + " (a int)"); err != nil {
		return fmt.Errorf("error creating table %s: %w", tableName, err)
	}

	for i := 0; i < 5; i++ {
		if _, err := db.Exec(fmt.Sprintf("insert into %s values (%d)", tableName, i)); err != nil {
			return fmt.Errorf("error inserting value %q into table %s: %w", i, tableName, err)
		}
	}

	return nil
}

// teardown ensure the table used for the example is removed.
func teardown(db *sql.DB) error {
	if _, err := db.Exec(fmt.Sprintf("if object_id('%s') is not null drop table %s", tableName, tableName)); err != nil {
		return fmt.Errorf("error dropping table %s: %w", tableName, err)
	}

	return nil
}

// example opens a cursor on the example table and removes rows with
// where the value 'a' is 3.
//
// This example is equivalent to executing the SQL statement
// 'drop from $table where a = 3'.
func example(driverConn interface{}) error {
	conn, ok := driverConn.(*ase.Conn)
	if !ok {
		return fmt.Errorf("driverConn %q is of type %T, not *ase.Conn", driverConn, driverConn)
	}

	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName)
	if err != nil {
		return fmt.Errorf("error opening cursor: %w", err)
	}
	defer cursor.Close(context.Background())

	rows, err := cursor.Fetch(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching result set: %w", err)
	}
	defer rows.Close()

	var rowCount = 0
	values := []driver.Value{0}

	for {
		if err := rows.Next(values); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("error reading row into values: %w", err)
		}

		fmt.Printf("row: %d a: %d\n", rowCount, values[0])
		rowCount += 1

		if values[0].(int32) == 3 {
			fmt.Println("dropping row")
			if err := rows.Delete(context.Background()); err != nil {
				return fmt.Errorf("error dropping row: %w", err)
			}
		}
	}

	return nil
}

func printTable(db *sql.DB) error {
	rows, err := db.Query("select * from " + tableName)
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to retrieve column names: %w", err)
	}

	fmt.Printf("| %s |\n", colNames[0])
	format := "| %d |\n"

	var a int

	for rows.Next() {
		if err = rows.Scan(&a); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Printf(format, a)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	return nil
}
