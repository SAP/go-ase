package tds

import (
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

func (tdsconn *TDSConn) Close() error {
	return tdsconn.conn.Close()
}
