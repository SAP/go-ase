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
	caps, err := NewCapabilityPackage(
		[]RequestCapability{
			// Support language requests
			TDS_REQ_LANG,
			// Support RPC requests
			// TODO: TDS_REQ_RPC,
			// Support procedure event notifications
			// TODO: TDS_REQ_EVT,
			// Support multiple commands per request
			TDS_REQ_MSTMT,
			// Support bulk copy
			// TODO: TDS_REQ_BCP,
			// Support cursors requests
			TDS_REQ_CURSOR,
			// Support dynamic SQL
			TDS_REQ_DYNF,
			// Support MSG requests
			TDS_REQ_MSG,
			// RPC will use TDS_DBRPC and TDS_PARAMFMT / TDS_PARAM
			TDS_REQ_PARAM,

			// Enable all optional data types
			TDS_DATA_INT1,
			TDS_DATA_INT2,
			TDS_DATA_INT4,
			TDS_DATA_BIT,
			TDS_DATA_CHAR,
			TDS_DATA_VCHAR,
			TDS_DATA_BIN,
			TDS_DATA_VBIN,
			TDS_DATA_MNY8,
			TDS_DATA_MNY4,
			TDS_DATA_DATE8,
			TDS_DATA_DATE4,
			TDS_DATA_FLT4,
			TDS_DATA_FLT8,
			TDS_DATA_NUM,
			TDS_DATA_TEXT,
			TDS_DATA_IMAGE,
			TDS_DATA_DEC,
			TDS_DATA_LCHAR,
			TDS_DATA_LBIN,
			TDS_DATA_INTN,
			TDS_DATA_DATETIMEN,
			TDS_DATA_MONEYN,
			TDS_DATA_SENSITIVITY,
			TDS_DATA_BOUNDARY,
			TDS_DATA_FLTN,
			TDS_DATA_BITN,
			TDS_DATA_INT8,
			TDS_DATA_UINT2,
			TDS_DATA_UINT4,
			TDS_DATA_UINT8,
			TDS_DATA_UINTN,
			TDS_DATA_NLBIN,
			TDS_IMAGE_NCHAR,
			TDS_BLOB_NCHAR_16,
			TDS_BLOB_NCHAR_8,
			TDS_BLOB_NCHAR_SCSU,
			TDS_DATA_DATE,
			TDS_DATA_TIME,
			TDS_DATA_INTERVAL,
			TDS_DATA_UNITEXT,
			TDS_DATA_SINT1,
			TDS_REQ_LARGEIDENT,
			TDS_REQ_BLOB_NCHAR_16,
			TDS_DATA_XML,
			TDS_DATA_BIGDATETIME,
			TDS_DATA_USECS,
			//TODO: TDS_DATA_LOBLOCATOR,

			// Support streaming
			//TODO: TDS_OBJECT_CHAR,
			//TODO: TDS_OBJECT_BINARY,

			// Support expedited and non-expedited attentions
			TDS_CON_OOB,
			TDS_CON_INBAND,
			// Use urgent notifications
			TDS_REQ_URGEVT,

			// Create procs from dynamic statements
			TDS_PROTO_DYNPROC,

			// Request status byte in TDS_PARAMS responses
			// Allows to handel nullbytes
			TDS_DATA_COLUMNSTATUS,
			// Support newer versions of tokens
			TDS_REQ_CURINFO3,
			TDS_REQ_DBRPC2,
			// TDS_PARAMFMT2
			TDS_WIDETABLES,

			// Support scrollable cursors
			TDS_CSR_SCROLL,
			TDS_CSR_SENSITIVE,
			TDS_CSR_INSENSITIVE,
			TDS_CSR_SEMISENSITIVE,
			TDS_CSR_KEYSETDRIVEN,

			// Renegotiate packet size after login negotiation
			//TODO: TDS_REQ_SRVPKTSIZE,

			// Support cluster failover and migration
			//TODO: TDS_CAP_CLUSTERFAILOVER,
			//TODO: TDS_REQ_MIGRATE,

			// Support batched parameters
			TDS_REQ_DYN_BATCH,
			TDS_REQ_LANG_BATCH,
			TDS_REQ_RPC_BATCH,

			// Support on demand encryption
			//TODO: TDS_REQ_COMMAND_ENCRYPTION,

			// Client will only perform readonly operations
			//TODO: TDS_REQ_READONLY,
		},
		[]ResponseCapability{
			// Ignore format control
			TDS_RES_NO_TDSCONTROL,
		},
		[]SecurityCapability{},
	)

	if err != nil {
		return fmt.Errorf("error creating capability package: %w", err)
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
