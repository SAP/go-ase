package libtest

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SAP/go-ase/libase/libdsn"
)

// SetupDB creates a database and sets .Database on the passed testDsn
func SetupDB(testDsn *libdsn.DsnInfo) error {
	db, err := sql.Open("ase", testDsn.AsSimple())
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to open connection: %v", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(context.Background(), "use master")
	if err != nil {
		return fmt.Errorf("Failed to switch context to master: %v", err)
	}

	testDatabase := "test" + RandomNumber()

	_, err = conn.ExecContext(context.Background(), "if db_id('?') is not null drop database ?", testDatabase, testDatabase)
	if err != nil {
		return fmt.Errorf("Error on conditional drop of database: %v", err)
	}

	_, err = conn.ExecContext(context.Background(), "create database ?", testDatabase)
	if err != nil {
		return fmt.Errorf("Failed to create database: %v", err)
	}

	testDsn.Database = testDatabase
	return nil
}

// TeardownDB deletes the database indicated by .Database of the passed
// testDsn and unsets the member.
func TeardownDB(testDsn *libdsn.DsnInfo) error {
	db, err := sql.Open("ase", testDsn.AsSimple())
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to open connection: %v", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(context.Background(), "use master")
	if err != nil {
		return fmt.Errorf("Failed to switch context to master: %v", err)
	}

	_, err = conn.ExecContext(context.Background(), "drop database ?", testDsn.Database)
	if err != nil {
		return fmt.Errorf("Failed to drop database: %v", err)
	}

	testDsn.Database = ""
	return nil
}

// SetupTableInsert creates a table with the passed type and inserts all
// passed samples as rows.
func SetupTableInsert(db *sql.DB, tableName, aseType string, samples ...interface{}) (*sql.Rows, func() error, error) {
	_, err := db.Exec("create table ? (a ?)", tableName, aseType)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create table: %v", err)
	}

	stmt, err := db.Prepare(fmt.Sprintf("insert into %s values (?)", tableName))
	if err != nil {
		return nil, nil, fmt.Errorf("Error preparing statement: %v", err)
	}
	defer stmt.Close()

	for _, sample := range samples {
		_, err := stmt.Exec(sample)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to execute prepared statement with %v: %v", sample, err)
		}
	}

	rows, err := db.Query("select * from ?", tableName)
	if err != nil {
		return nil, nil, fmt.Errorf("Error selecting from %s: %v", tableName, err)
	}

	teardownFn := func() error {
		_, err := db.Exec("drop table ?", tableName)
		return err
	}

	return rows, teardownFn, nil
}
