package tds

import (
	"context"
	"fmt"
	"io"
)

type Message struct {
	packages []Package
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

	// errCh will receive errors from the goroutines and exit
	errCh := make(chan error, 1)
	defer func() { close(errCh) }()

	// ctx singals goroutines to exit their event loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Split input into packets.
	// A size of two should be enough to ensure consistent throughput.
	// TODO channel size configurable? Depending on the PDU size
	// (default 512b, max 32767b -> 32Mb), which can put significant
	// strain on the client machine with multiplexed connections
	packetCh := make(chan *Packet, 2)
	go msg.readFromPackets(ctx, errCh, reader, packetCh)

	packageCh := make(chan Package, 1)
	go msg.readFromPackages(ctx, errCh, packetCh, packageCh)

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

func (msg *Message) readFromPackets(ctx context.Context, errCh chan error, reader io.Reader, packetCh chan *Packet) {
	defer close(packetCh)

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

		packetCh <- packet

		if packet.Header.Status == TDS_BUFSTAT_EOM {
			return
		}
	}
}

func (msg *Message) readFromPackages(ctx context.Context, errCh chan error, packetCh chan *Packet, packageCh chan Package) {
	defer close(packageCh)

	// package.Data is written to byteCh for packages to retrieve
	// from
	byteCh := newChannel()
	defer func() { byteCh.Close() }()

	var pkg Package

	for {
		select {
		case <-ctx.Done():
			return
		case packet, ok := <-packetCh:
			if !ok {
				packetCh = nil
				continue
			}

			bs := packet.Data

			// Create new package if pkg is currently unset
			if pkg == nil {
				// retrieve and validate TDSToken
				token := (TDSToken)(bs[0])

				// cut off token from data
				bs = bs[1:]

				// create new package based on token
				pkg, ok = tokenToPackage[token]
				if !ok {
					errCh <- fmt.Errorf("no package found for token: %s", token)
					return
				}

				// Start goroutine reading from byte channel
				go pkg.ReadFrom(byteCh)
			}

			// Write bytes to bytechannel
			byteCh.WriteBytes(bs)
		default:
		}

		// Check if current package is finished
		if pkg != nil && pkg.Finished() {
			if err := pkg.Error(); err != nil {
				errCh <- fmt.Errorf("error ocurred while parsing packet into package: %v", err)
				return
			}

			// Send package to channel and set for next loop
			packageCh <- pkg
			pkg = nil
		}

		if packetCh == nil && pkg == nil {
			return
		}
	}
}
