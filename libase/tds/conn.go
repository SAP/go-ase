package tds

import (
	"fmt"
	"net"
)

// TDSConn handles a TDS-based connection.
type TDSConn struct {
	conn net.Conn
}

func Dial(network, address string) (*TDSConn, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &TDSConn{conn: c}, nil
}

func (tds *TDSConn) Close() error {
	return tds.conn.Close()
}

func (tds *TDSConn) Receive() (*Message, error) {
	msg := &Message{}

	err := msg.ReadFrom(tds.conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %v", err)
	}

	return msg, nil
}

// Send transmits a messages payload to the server.
func (tds *TDSConn) Send(msg Message) error {
	return msg.WriteTo(tds.conn)
}
