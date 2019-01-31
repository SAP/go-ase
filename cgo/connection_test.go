package cgo

import (
	"context"
	"database/sql/driver"
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
