package tds

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

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

			log.Printf("package: %#v", pkg)
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

	n := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packet := &Packet{}
		_, err := packet.ReadFrom(reader)
		if err != nil {
			errCh <- fmt.Errorf("error reading packet: %v", err)
			return
		}

		byteCh.WriteBytes(packet.Data)
		err = ioutil.WriteFile(fmt.Sprintf("/sybase/TST/packet-%d.bin", n), packet.Data, 0660)
		n++
		if err != nil {
			log.Printf("err writing file: %v", err)
		}

		if packet.Header.Status == TDS_BUFSTAT_EOM {
			return
		}
	}
}

type LastPkgAcceptor interface {
	LastPkg(Package) error
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
			if err == ErrChannelExhausted {
				continue
			}
			if err == io.EOF {
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

		// Start goroutine reading from byte channel
		if err := pkg.ReadFrom(byteCh); err != nil {
			log.Printf(">>> Error from pkg.ReadFrom: %v", err)
			errCh <- fmt.Errorf("error ocurred while parsing packet into package: %v", err)
			return
		}

		packageCh <- pkg
		lastpkg = pkg
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

			packet.Header.MsgType = msg.headerType

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
		log.Printf("package: %#v", pack)
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
		if err != nil && err != io.EOF {
			errCh <- fmt.Errorf("error occurred reading data into packet: %w", err)
			return
		}

		packet.Header.Length = uint16(n + MsgHeaderLength)
		packet.Data = data[:n]

		if err == io.EOF {
			packet.Header.Status |= TDS_BUFSTAT_EOM
		}

		packetCh <- packet

		if err == io.EOF {
			return
		}
	}
}
