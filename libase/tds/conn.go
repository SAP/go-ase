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
	caps *CapabilityPackage
}

func Dial(network, address string) (*TDSConn, error) {
	tds := &TDSConn{}

	err := tds.setCapabilities()
	if err != nil {
		return nil, fmt.Errorf("error setting capabilities on connection: %w", err)
	}

	c, err := net.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}

	tds.conn = c
	return tds, nil
}

func (tds *TDSConn) setCapabilities() error {
	caps := NewCapabilityPackage()

	// Request status byte in TDS_PARAMS responses
	// Allows to handel nullbytes
	err := caps.SetRequestCapability(TDS_DATA_COLUMNSTATUS, true)
	if err != nil {
		return fmt.Errorf("failed to set request capability %s: %w", TDS_DATA_COLUMNSTATUS, err)
	}

	// Signal ability to handle TDS_PARAMFMT2
	err = caps.SetResponseCapability(TDS_WIDETABLES, true)
	if err != nil {
		return fmt.Errorf("failed to set response capability %s: %w", TDS_WIDETABLES, err)
	}

	tds.caps = caps
	return nil
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
