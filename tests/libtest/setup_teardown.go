package libtests

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SAP/go-ase/libase/dsn"
)

// SetupDB creates a database and sets .Database on the passed testDsn
func SetupDB(testDsn *dsn.DsnInfo) error {
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
func TeardownDB(testDsn *dsn.DsnInfo) error {
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
