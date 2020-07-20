package tds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
)

// TDSChannel is a channel in a multiplexed connection with a TDS
// server.
type TDSChannel struct {
	tdsConn *TDSConn

	channelId          int
	envChangeHooks     []EnvChangeHook
	envChangeHooksLock *sync.Mutex

	// currentHeaderType is the MessageHeaderType set on outgoing
	// packets.
	CurrentHeaderType MessageHeaderType
	// curPacketNr is the number of the next packet being sent
	curPacketNr int
	// window is the amount of buffers transmitted between ACKs
	window int

	// packets stores unconsumed Packets
	queue *PacketQueue
	// packageCh stores Packages as they are parsed from Packets
	packageCh chan Package
	// lastPkg is the last package sent by the TDS server and added to
	// the channel.
	lastPkg Package

	errCh chan error
}

// NewTDSChannel communicates the creation of a new channel with the
// server.
func (tds *TDSConn) NewTDSChannel(packageChannelSize int) (*TDSChannel, error) {
	channelId, err := tds.getValidChannelId()
	if err != nil {
		return nil, fmt.Errorf("error getting channel ID: %w", err)
	}

	tdsChan := &TDSChannel{
		tdsConn:           tds,
		channelId:         channelId,
		envChangeHooksLock: &sync.Mutex{},
		CurrentHeaderType: TDS_BUF_NORMAL,
		window:            100, // TODO
		queue:             NewPacketQueue(),
		packageCh:         make(chan Package, packageChannelSize),
		errCh:             make(chan error, 10),
	}

	tds.tdsChannels[channelId] = tdsChan

	// Send packets to setup logical channel
	setup := NewPacket()
	setup.Header.Length = MsgHeaderLength
	setup.Data = nil

	tdsChan.CurrentHeaderType = TDS_BUF_SETUP
	err = tdsChan.sendPacket(setup)
	if err != nil {
		return nil, fmt.Errorf("error sending setup for channel %d: %w",
			tdsChan.channelId, err)
	}

	pkg, err := tdsChan.NextPackage(true)
	if err != nil {
		return nil, fmt.Errorf("error receiving ack for channel setup: %w", err)
	}

	header, ok := pkg.(*HeaderOnlyPackage)
	if !ok {
		return nil, fmt.Errorf("did not received expected header-only packet: %v", pkg)
	}

	if header.Header.MsgType&TDS_BUF_PROTACK != TDS_BUF_PROTACK {
		return nil, fmt.Errorf("did not receive protack in header-only packet: %s",
			header)
	}

	tdsChan.ResetHeaderType()
	return tdsChan, nil
}

// Reset resets the TDSChannel after a communication has been completed.
func (tdsChan *TDSChannel) Reset() {
	tdsChan.CurrentHeaderType = TDS_BUF_NORMAL
	tdsChan.queue.Reset()
	tdsChan.lastPkg = nil
}

// Close communicates the closing of the channel with the TDS server.
//
// The teardown on client side is guaranteed, even if Close returns an
// error. An error is only returned if the communication with the server
// fails.
func (tdsChan *TDSChannel) Close() error {
	defer close(tdsChan.packageCh)
	defer close(tdsChan.errCh)

	// Send packet to teardown logical channel
	teardown := NewPacket()
	teardown.Header.Length = MsgHeaderLength
	teardown.Data = nil
	tdsChan.CurrentHeaderType = TDS_BUF_CLOSE

	err := tdsChan.sendPacket(teardown)
	if err != nil {
		return fmt.Errorf("error sending teardown for channel %d: %w",
			tdsChan.channelId, err)
	}

	return nil
}

// Error returns either communications errors from the underlying
// TDSConn or parse errors from the TDSChannel.
func (tdsChan *TDSChannel) Error() error {
	if err, ok := tdsChan.tdsConn.Error(); ok {
		return fmt.Errorf("error in tds connection: %w", err)
	}

	if err, ok := <-tdsChan.errCh; ok {
		return fmt.Errorf("error in tds channel: %w", err)
	}

	return nil
}

// handleSpecialPackage handles special packages such as env changes.
// The returned boolean signals if the package should be passed along or
// skipped.
// An error is returned if the handling errored.
func (tdsChan *TDSChannel) handleSpecialPackage(pkg Package) (bool, error) {
	if envChange, ok := pkg.(*EnvChangePackage); ok {
		for _, member := range envChange.members {
			go tdsChan.callEnvChangeHooks(member.Type, member.NewValue, member.OldValue)
		}
		return false, nil
	}

	return true, nil
}

// NextPackage returns the next package in the queue.
// If wait is true an error is only returned when the context of the
// associated TDSConn is cancelled.
// If wait is false an error is only returned if no package is ready to
// be returned.
func (tdsChan *TDSChannel) NextPackage(wait bool) (Package, error) {
	if wait {
		select {
		case <-tdsChan.tdsConn.ctx.Done():
			return nil, errors.New("context cancelled")
		case pkg := <-tdsChan.packageCh:
			return pkg, nil
		}
	}

	pkg, ok := <-tdsChan.packageCh
	if !ok {
		return nil, errors.New("No Package ready")
	}

	return pkg, nil
}

// AddPackage utilized PacketQueue to convert a Package into packets.
// Packets that have their Data exhausted are sent to the server.
func (tdsChan *TDSChannel) AddPackage(pkg Package) error {
	if acceptor, ok := pkg.(LastPkgAcceptor); ok {
		err := acceptor.LastPkg(tdsChan.lastPkg)
		if err != nil {
			return fmt.Errorf("error calling LastPkg: %w", err)
		}
	}

	err := pkg.WriteTo(tdsChan.queue)
	if err != nil {
		return fmt.Errorf("error queueing packets from package: %w", err)
	}
	tdsChan.lastPkg = pkg

	return tdsChan.sendPackets(true)
}

// Send all remaining Packets in queue to the server.
// This includes Packets whose Data isn't exhausted.
func (tdsChan *TDSChannel) SendRemainingPackets() error {
	return tdsChan.sendPackets(false)
}

func (tdsChan *TDSChannel) sendPackets(onlyFull bool) error {
	defer func() { tdsChan.queue.DiscardUntilCurrentPosition() }()

	for i, packet := range tdsChan.queue.queue {
		// Only the last packet should not be full.
		if i == tdsChan.queue.indexPacket && tdsChan.queue.indexData < MsgBodyLength {
			if onlyFull {
				// Packet is not exhausted and only exhausted packets
				// should be sent. Return.
				return nil
			}

			// Packet is not exhausted but should be sent. Adjust header
			// length
			packet.Header.Length = uint16(MsgHeaderLength + tdsChan.queue.indexData)
			packet.Data = packet.Data[:tdsChan.queue.indexData]
		}

		// TODO maybe check if data is empty - could be an issue

		err := tdsChan.sendPacket(packet)
		if err != nil {
			return fmt.Errorf("error sending packet: %w", err)
		}
	}

	return nil
}

func (tdsChan *TDSChannel) sendPacket(packet *Packet) error {
	packet.Header.MsgType = tdsChan.CurrentHeaderType
	packet.Header.Channel = uint16(tdsChan.channelId)
	packet.Header.PacketNr = uint8(tdsChan.curPacketNr)
	tdsChan.curPacketNr = (tdsChan.curPacketNr + 1) % 256
	packet.Header.Window = uint8(tdsChan.window)

	if len(packet.Data) != MsgBodyLength {
		// Data portion is not exhausted, this is the last packet.
		packet.Header.Status |= TDS_BUFSTAT_EOM
	}

	n, err := packet.WriteTo(tdsChan.tdsConn.conn)
	if err != nil {
		return fmt.Errorf("error writing packet to server: %w", err)
	}

	if int(n) != int(packet.Header.Length) {
		return fmt.Errorf("expected to write %d bytes for packet, wrote %d instead",
			int(packet.Header.Length)+MsgHeaderLength, n)
	}

	return nil
}

// WritePacket received packets from the associated TDSConn and attempts
// to produce Packages from the existing data.
func (tdsChan *TDSChannel) WritePacket(packet *Packet) {
	// The packet is header-only - pass it directly into the package
	// channel.
	if packet.Header.Length == MsgHeaderLength {
		tdsChan.packageCh <- HeaderOnlyPackage{Header: packet.Header}
		return
	}

	// Add packet into queue
	tdsChan.queue.AddPacket(packet)

	for {
		// Read out current position for resetting if the existing data
		// isn't enough to fill a Package.
		curPacket, curData := tdsChan.queue.Position()

		// Attempt to parse a Package.
		ok := tdsChan.tryParsePackage()
		if !ok {
			// Attempt failed, roll back position and return.
			tdsChan.queue.SetPosition(curPacket, curData)
			return
		}

		// Package could be filled with the available data. Discard all
		// consumed packets.
		tdsChan.queue.DiscardUntilCurrentPosition()
	}
}

// tryParsePackage attempts to parse a Package from the queued Packets.
func (tdsChan *TDSChannel) tryParsePackage() bool {
	// Attempt to process data from channel into a Package.
	tokenByte, err := tdsChan.queue.Byte()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			tdsChan.errCh <- fmt.Errorf("error reading token byte: %w", err)
		}
		return false
	}

	// Create Package.
	pkg, err := LookupPackage(TDSToken(tokenByte))
	if err != nil {
		tdsChan.errCh <- err
		return false
	}

	// If the Package is tokenless write the token byte back in.
	if tokenless, ok := pkg.(*TokenlessPackage); ok {
		tokenless.Data.WriteByte(tokenByte)
	}

	// Read data into Package.
	err = pkg.ReadFrom(tdsChan.queue)
	if err != nil {
		// Not enough data available
		// TODO: create an explicit error to check for
		return false
	}

	pass, err := tdsChan.handleSpecialPackage(pkg)
	if err != nil {
		tdsChan.errCh <- fmt.Errorf("error while handling sepcial package: %w", err)
		// Package handling errored, but the package could be parsed.
		// Continue.
		return true
	}

	if !pass {
		// Package should not be handled further, continue
		return true
	}

	tdsChan.packageCh <- pkg
	tdsChan.lastPkg = pkg
	return true
}
