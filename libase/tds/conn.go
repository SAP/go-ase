package tds

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// TDSConn handles a TDS-based connection.
type TDSConn struct {
	conn               io.ReadWriteCloser
	caps               *CapabilityPackage
	envChangeHooks     []EnvChangeHook
	envChangeHooksLock *sync.Mutex
}

func Dial(network, address string) (*TDSConn, error) {
	tds := &TDSConn{}

	err := tds.setCapabilities()
	if err != nil {
		return nil, fmt.Errorf("error setting capabilities on connection: %w", err)
	}

	tds.envChangeHooksLock = &sync.Mutex{}

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
	err = caps.SetRequestCapability(TDS_WIDETABLES, true)
	if err != nil {
		return fmt.Errorf("failed to set request capability %s: %w", TDS_WIDETABLES, err)
	}

	tds.caps = caps
	return nil
}

// RegisterEnvChangeHook register functions of the type EnvChangeHook.
//
// The registered functions are called with the EnvChangeType of the
// update, the old value and the new value.
func (tds *TDSConn) RegisterEnvChangeHook(fn EnvChangeHook) {
	tds.envChangeHooksLock.Lock()
	defer tds.envChangeHooksLock.Unlock()

	tds.envChangeHooks = append(tds.envChangeHooks, fn)
}

// TODO when to call this?
// possible would be as the data stream is parsed
// also possible would be after an entire data stream has been processed
func (tds *TDSConn) callEnvChangeHooks(typ EnvChangeType, oldValue, newValue string) {
	tds.envChangeHooksLock.Lock()
	defer tds.envChangeHooksLock.Unlock()

	for _, fn := range tds.envChangeHooks {
		fn(typ, oldValue, newValue)
	}
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
