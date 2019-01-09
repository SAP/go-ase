package cgo

import "testing"

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
