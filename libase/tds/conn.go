package tds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/hashicorp/go-multierror"
)

// Conn handles a TDS-based connection.
//
// Note: This is not the underlying structure for driver.Conn - that is
// Channel.
type Conn struct {
	conn io.ReadWriteCloser
	caps *CapabilityPackage
	dsn  *libdsn.Info

	odce odceCipher

	ctx                 context.Context
	ctxCancel           context.CancelFunc
	tdsChannelCurFreeId uint32
	tdsChannels         map[int]*Channel
	tdsChannelsLock     *sync.RWMutex
	errCh               chan error
}

// Dial returns a prepared and dialed Conn.
func NewConn(ctx context.Context, dsn *libdsn.Info) (*Conn, error) {
	network := "tcp"
	if prop := dsn.Prop("network"); prop != "" {
		network = prop
	}

	c, err := net.Dial(network, fmt.Sprintf("%s:%s", dsn.Host, dsn.Port))
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}

	tds := &Conn{}
	tds.dsn = dsn
	tds.conn = c

	err = tds.setCapabilities()
	if err != nil {
		return nil, fmt.Errorf("error setting capabilities on connection: %w", err)
	}

	tds.odce = aes_256_cbc

	tds.ctx, tds.ctxCancel = context.WithCancel(ctx)
	// Channels cannot have ID 0 - but channel with the id 0 is used to
	// communicate general packets such as login/logout.
	tds.tdsChannelCurFreeId = uint32(0)
	tds.tdsChannels = make(map[int]*Channel)
	tds.tdsChannelsLock = &sync.RWMutex{}
	tds.errCh = make(chan error, 10)

	// A goroutine automatically reads payloads from the server and
	// passes them to the corresponding channel.
	// Payloads sent to the server are sent in the thread the client
	// uses.
	go tds.ReadFrom()

	return tds, nil
}

// Close closes a Conn and its unclosed Channels.
//
// Teardown and closing on the client side is guaranteed, even if Close
// returns an error. An error is only returned if the communication with
// the server fails or if channels report errors during closing.
//
// If an error is returned it is a *multierror.Error with all errors.
func (tds *Conn) Close() error {
	tds.ctxCancel()

	var me error

	for _, channel := range tds.tdsChannels {
		err := channel.Close()
		if err != nil {
			me = multierror.Append(me, fmt.Errorf("error closing channel: %w", err))
		}
	}

	err := tds.conn.Close()
	if err != nil {
		me = multierror.Append(me, fmt.Errorf("error closing connection: %w", err))
	}

	return me
}

func (tds *Conn) getValidChannelId() (int, error) {
	curId := int(tds.tdsChannelCurFreeId)

	if curId > math.MaxUint16 {
		// TODO create error
		return 0, fmt.Errorf("exhausted all channel IDs")
	}

	// increment ID before recursing or returning
	atomic.AddUint32(&tds.tdsChannelCurFreeId, 1)

	if _, ok := tds.tdsChannels[curId]; ok {
		// ChannelId is already used, recurse
		return tds.getValidChannelId()
	}

	return curId, nil
}

// ReadFrom creates packets from payloads from the server and writes
// them to the corresponding Channel.
func (tds *Conn) ReadFrom() {
	for {
		select {
		case <-tds.ctx.Done():
			return
		default:
			packet := &Packet{}
			_, err := packet.ReadFrom(tds.conn)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				tds.errCh <- fmt.Errorf("error reading packet: %w", err)
				continue
			}

			tds.tdsChannelsLock.RLock()
			tdsChan, ok := tds.tdsChannels[int(packet.Header.Channel)]
			tds.tdsChannelsLock.RUnlock()
			if !ok {
				tds.errCh <- fmt.Errorf("received packet for invalid channel %d", packet.Header.Channel)
				continue
			}

			// Errors are recorded in the channels' error channel.
			tdsChan.WritePacket(packet)
		}
	}
}

func (tds *Conn) setCapabilities() error {
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
			// TODO: TDS_REQ_CURSOR,
			// Support dynamic SQL
			// TODO: TDS_REQ_DYNF,
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
			TDS_REQ_COMMAND_ENCRYPTION,

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
