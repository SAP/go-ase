package tds

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type LastPkgAcceptor interface {
	LastPkg(Package) error
}

type Message struct {
	headerType MessageHeaderType
	packages   []Package
}

func NewMessage() *Message {
	return &Message{}
}

func (msg Message) Packages() []Package {
	return msg.packages
}

func (msg *Message) AddPackage(pack Package) {
	if acceptor, ok := pack.(LastPkgAcceptor); ok {
		acceptor.LastPkg(msg.packages[len(msg.packages)-1])
	}
	msg.packages = append(msg.packages, pack)
}

func (msg *Message) ReadFrom(reader io.Reader) error {
	// TODO canceling during processing w/ TDS_BUFSTAT_ATTN
	// goroutines with sync?
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// errCh will receive errors from the goroutines and exit
	errCh := make(chan error, 1)
	defer close(errCh)

	byteCh := newChannel()

	// Split input into packets and write the bodies into the byte
	// channel
	go msg.readFromPackets(ctx, errCh, reader, byteCh)

	packageCh := make(chan Package, 1)
	go msg.readFromPackages(ctx, errCh, byteCh, packageCh)

	for {
		select {
		case err := <-errCh:
			return err
		case pkg, ok := <-packageCh:
			if !ok {
				packageCh = nil
				continue
			}

			msg.AddPackage(pkg)
		default:
		}

		if packageCh == nil {
			break
		}
	}

	return nil
}

func (msg *Message) readFromPackets(ctx context.Context, errCh chan error, reader io.Reader, byteCh *channel) {
	defer byteCh.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packet := &Packet{}
		_, err := packet.ReadFrom(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			} else {
				errCh <- fmt.Errorf("error reading packet: %w", err)
				return
			}
		}

		byteCh.WriteBytes(packet.Data)

		if packet.Header.Status == TDS_BUFSTAT_EOM {
			return
		}
	}
}

func (msg *Message) readFromPackages(ctx context.Context, errCh chan error, byteCh *channel, packageCh chan Package) {
	defer close(packageCh)

	var lastpkg Package
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		tokenByte, err := byteCh.Byte()
		if err != nil {
			if errors.Is(err, ErrChannelExhausted) {
				continue
			}
			if errors.Is(err, io.EOF) {
				return
			}
			errCh <- fmt.Errorf("error reading token byte from channel: %w", err)
			return
		}

		pkg, err := LookupPackage(TDSToken(tokenByte))
		if err != nil {
			errCh <- err
			return
		}

		if acceptor, ok := pkg.(LastPkgAcceptor); ok {
			err := acceptor.LastPkg(lastpkg)
			if err != nil {
				errCh <- fmt.Errorf("error reading information from last package: %w", err)
				return
			}
		}

		// Write tokenByte into the tokenless package data since the
		// received package isn't token-based
		if tokenless, ok := pkg.(*TokenlessPackage); ok {
			tokenless.Data.WriteByte(tokenByte)
		}

		// Start goroutine reading from byte channel
		if err := pkg.ReadFrom(byteCh); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			errCh <- fmt.Errorf("error ocurred while parsing packet into package: %w", err)
			return
		}

		packageCh <- pkg
		lastpkg = pkg

		if IsDone(lastpkg) {
			return
		}
	}
}

func (msg Message) WriteTo(writer io.Writer) error {
	errCh := make(chan error, 1)
	defer close(errCh)

	byteCh := newChannel()

	go msg.writeToPackage(errCh, byteCh)

	packetCh := make(chan Packet, 1)
	go msg.writeToPackets(errCh, byteCh, packetCh)

	for {
		select {
		case err := <-errCh:
			return err
		case packet, ok := <-packetCh:
			if !ok {
				return nil
			}

			// Assume TDS_BUF_NORMAL for message type unless it was
			// explicitly set.
			if msg.headerType == 0x0 {
				packet.Header.MsgType = TDS_BUF_NORMAL
			} else {
				packet.Header.MsgType = msg.headerType
			}

			_, err := packet.WriteTo(writer)
			if err != nil {
				return fmt.Errorf("error occurred while writing packet to writer: %w", err)
			}
		}
	}
}

func (msg Message) writeToPackage(errCh chan error, byteCh *channel) {
	defer byteCh.Close()
	for _, pack := range msg.Packages() {
		err := pack.WriteTo(byteCh)
		if err != nil {
			errCh <- fmt.Errorf("failed to write package data to channel: %w", err)
			return
		}
	}
}

func (msg Message) writeToPackets(errCh chan error, byteCh *channel, packetCh chan Packet) {
	defer close(packetCh)

	for {
		packet := Packet{}
		packet.Header = MessageHeader{}
		data := make([]byte, MsgBodyLength)

		n, err := byteCh.Read(data)
		if err != nil && !errors.Is(err, io.EOF) {
			errCh <- fmt.Errorf("error occurred reading data into packet: %w", err)
			return
		}

		packet.Header.Length = uint16(n + MsgHeaderLength)
		packet.Data = data[:n]

		if errors.Is(err, io.EOF) {
			packet.Header.Status |= TDS_BUFSTAT_EOM
		}

		packetCh <- packet

		// Exit if the header signal end of message
		if packet.Header.Status&TDS_BUFSTAT_EOM == TDS_BUFSTAT_EOM {
			return
		}
	}
}
