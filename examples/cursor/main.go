// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to use cursor with the go-ase driver.
//
// go-ase uses cursors for queries by default, but that can be disabled
// making creating cursors through the driver directly the only option
// of using cursors.
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
	"github.com/SAP/go-ase/examples"
	"github.com/SAP/go-dblib/dsn"
)

const (
	exampleName  = "cursor"
	databaseName = exampleName + "DB"
	tableName    = databaseName + ".." + exampleName + "Table"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("%s failed: %v", exampleName, err)
	}
}

func DoMain() error {
	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	dropDB, err := examples.CreateDropDatabase(info, databaseName)
	if err != nil {
		return err
	}
	defer dropDB()

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer db.Close()

	dropTable, err := examples.CreateDropTable(db, tableName, "a int, b varchar(30)")
	if err != nil {
		return err
	}
	defer dropTable()

	return Test(db)
}

func Test(db *sql.DB) error {
	for i, val := range []string{"one", "two", "three"} {
		if _, err := db.Exec("insert into "+tableName+" (a, b) values (?, ?)", i+1, val); err != nil {
			return fmt.Errorf("failed to insert values: %w", err)
		}
	}

	// database/sql doesn't have a cursor interface, instead the
	// underlying go-ase Conn must be used.
	//
	// This is achieved by retrieving a database/sql.Conn from the
	// connection pool and then utilizing the go-ase.Conn through
	// database/sql.Conn.Raw.
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

	// A cursor is opened explicitly through the .NewCursor command,
	// after which a CursorRows can be retrieved through .Fetch.
	//
	// Note that the CursorRows implements the database/sql/driver.Rows
	// interface and does not reflect the signature of
	// database/sql.Rows.
	fmt.Println("opening cursor1")
	cursor, err := conn.NewCursor(context.Background(), "select * from "+tableName)
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
			log.Printf("error closing rows of cursor1: %v", err)
		}
	}()

	fmt.Println("iterating over cursor1")
	values := []driver.Value{0, ""}
	for {
		// Contrary to database/sql.Rows the .Next method returns an error
		// rather than a boolean.
		if err := rows.Next(values); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading row: %w", err)
		}

		fmt.Printf("a: %d\n", values[0])
		fmt.Printf("b: %s\n", values[1])
	}

	fmt.Println("opening cursor2")
	// Creating a cursor with arguments and its handling is equivalent
	// to a cursor without arguments.
	cursor2, err := conn.NewCursor(context.Background(),
		"select a, b from "+tableName+" where b like (?)", "two")
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
			log.Printf("error closing rows of cursor2: %v", err)
		}
	}()

	fmt.Println("iterating over cursor2")
	values2 := []driver.Value{0, ""}
	for {
		if err := rows2.Next(values2); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading row: %w", err)
		}

		fmt.Printf("a: %d\n", values2[0])
		fmt.Printf("b: %s\n", values2[1])
	}

	return nil
}
