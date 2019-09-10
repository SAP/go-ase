package tds

import (
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

	byteCh := newChannel()

	// Split input into packets and write the bodies into the byte
	// channel
	go msg.readFromPackets(errCh, reader, byteCh)

	packageCh := make(chan Package, 1)
	go msg.readFromPackages(errCh, byteCh, packageCh)

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

func (msg *Message) readFromPackets(errCh chan error, reader io.Reader, byteCh *channel) {
	defer byteCh.Close()

	for {
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

func (msg *Message) readFromPackages(errCh chan error, byteCh *channel, packageCh chan Package) {
	defer close(packageCh)

	var lastpkg Package
	for {
		tokenByte, err := byteCh.Byte()
		if err != nil {
			if err == ErrChannelExhausted {
				continue
			}
			if err == ErrChannelClosed {
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
