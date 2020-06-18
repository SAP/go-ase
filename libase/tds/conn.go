package tds

import (
	"fmt"
	"io"
	"io/ioutil"
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
		return nil, fmt.Errorf("failed to read message: %v", err)
	}

	return msg, nil
}

// Send transmits a messages payload to the server.
func (tds *TDSConn) Send(msg Message) error {
	log.Printf("Sending message: %d Packages", len(msg.packages))

	for i, pack := range msg.packages {
		if ms, ok := pack.(MultiStringer); ok {
			for _, s := range ms.MultiString() {
				log.Printf("  %s", s)
			}
		} else {
			log.Printf("  Package %d: %s", i, pack)
			log.Printf("    %#v", pack)
		}

		stdoutCh := newChannel()
		pack.WriteTo(stdoutCh)
		stdoutCh.Close()
		bs, _ := ioutil.ReadAll(stdoutCh)
		log.Printf("    Bytes(%d): %#v", len(bs), bs)
	}

	return msg.WriteTo(tds.conn)
}
