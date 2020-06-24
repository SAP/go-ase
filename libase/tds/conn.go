package tds

import (
	"fmt"
	"io"
	"log"
	"net"
)

// TDSConn handles a TDS-based connection.
type TDSConn struct {
	conn io.ReadWriteCloser
}

func Dial(network, address string) (*TDSConn, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}

	return &TDSConn{conn: c}, nil
}

func (tds *TDSConn) Close() error {
	return tds.conn.Close()
}

type MultiStringer interface {
	MultiString() []string
}

func (tds *TDSConn) Receive() (*Message, error) {
	msg := &Message{}

	err := msg.ReadFrom(tds.conn)

	// TODO remove
	log.Printf("Received message: %d Packages", len(msg.packages))
	for i, pack := range msg.packages {
		if ms, ok := pack.(MultiStringer); ok {
			for _, s := range ms.MultiString() {
				log.Printf("  %s", s)
			}
		} else {
			log.Printf("  Package %d: %s", i, pack)
			log.Printf("    %#v", pack)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	return msg, nil
}

// Send transmits a messages payload to the server.
func (tds *TDSConn) Send(msg Message) error {
	log.Printf("Sending message: %d Packages", len(msg.packages))

	// TODO remove
	for i, pack := range msg.packages {
		log.Printf("  Package %d: %s", i, pack)
		if ms, ok := pack.(MultiStringer); ok {
			for _, s := range ms.MultiString() {
				log.Printf("    %s", s)
			}
		} else {
			log.Printf("    %#v", pack)
		}
	}

	return msg.WriteTo(tds.conn)
}
