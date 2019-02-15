package cgo

import (
	"context"
	"database/sql/driver"
	"math/rand"
	"strconv"
	"testing"
)

func TestNewConnection(t *testing.T) {
	conn, err := newConnection(*testDsn)
	if err != nil {
		t.Errorf("Unexpected error opening connection: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Errorf("Unexpected error closing connection: %v", err)
	}
}

func TestPing(t *testing.T) {
	conn, err := newConnection(*testDsn)
	if err != nil {
		t.Errorf("Unexpected error opening connection: %v", err)
	}
	defer conn.Close()

	err = conn.Ping(context.Background())
	if err != nil {
		t.Errorf("Got error pinging ASE: %v", err)
	}
}

func TestPingErr(t *testing.T) {
	conn, err := newConnection(*testDsn)
	if err != nil {
		t.Errorf("Unexpected error opening connection: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Errorf("Unexpected error closing connection: %v", err)
	}

	err = conn.Ping(context.Background())
	if err != driver.ErrBadConn {
		t.Errorf("Did not receive driver.ErrBadConn: %v", err)
	}
}

func TestExec(t *testing.T) {
	tableName := "table" + strconv.Itoa(rand.Int())

	conn, err := newConnection(*testDsn)
	if err != nil {
		t.Errorf("Unexpected error opening connection: %v", err)
		return
	}
	defer conn.Close()
	result, err := conn.Exec("create table "+tableName+" (a int, b char(3))", nil)

	if err != nil {
		t.Errorf("Received error when creating table: %v", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Errorf("Received error when reading affected rows: %v", err)
	} else {
		if rowsAffected != -1 {
			t.Errorf("Result indicated affected rows when there should be none: %d", rowsAffected)
		}
	}

	result, err = conn.Exec("drop table "+tableName, nil)
	if err != nil {
		t.Errorf("Received error when dropping table: %v", err)
	}
}
