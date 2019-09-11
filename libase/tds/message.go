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
			errCh <- fmt.Errorf("error reading packet: %v", err)
			return
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

		// Start goroutine reading from byte channel
		pkg.ReadFrom(byteCh)

		if err := pkg.Error(); err != nil {
			errCh <- fmt.Errorf("error ocurred while parsing packet into package: %v", err)
			return
		}

		packageCh <- pkg
		lastpkg = pkg
	}
}
