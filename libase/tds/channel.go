package tds

import (
	"errors"
	"fmt"
)

// TDSChannel is a channel in a multiplexed connection with a TDS
// server.
type TDSChannel struct {
	tdsConn *TDSConn

	channelId int

	// currentHeaderType is the MessageHeaderType set on outgoing
	// packets.
	currentHeaderType MessageHeaderType
	// packets stores unconsumed Packets
	queue *PacketQueue
	// packageCh stores Packages as they are parsed from Packets
	packageCh chan Package

	errCh chan error
}

func (tds *TDSConn) NewTDSChannel(packageChannelSize int) (*TDSChannel, error) {
	channelId, err := tds.getValidChannelId()
	if err != nil {
		return nil, fmt.Errorf("error getting channel ID: %w", err)
	}

	tdsChan := &TDSChannel{
		tdsConn:   tds,
		channelId: channelId,
		queue:     NewPacketQueue(),
		packageCh: make(chan Package, packageChannelSize),
		errCh:     make(chan error, 10),
	}

	tds.tdsChannels[channelId] = tdsChan

	return tdsChan, nil
}

func (tdsChan *TDSChannel) Close() {
	close(tdsChan.packageCh)
	close(tdsChan.errCh)
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
	err := pkg.WriteTo(tdsChan.queue)
	if err != nil {
		return fmt.Errorf("error queueing packets from package: %w", err)
	}

	return tdsChan.sendPackets(true)
}

// Send all remaining Packets in queue to the server.
// This includes Packets whose Data isn't exhausted.
func (tdsChan *TDSChannel) SendRemainingPackets() error {
	return tdsChan.sendPackets(false)
}

func (tdsChan *TDSChannel) sendPackets(onlyFull bool) error {
	defer func() { tdsChan.queue.DiscardUntilCurrentPosition() }()

	for _, packet := range tdsChan.queue.queue {
		// Check if Packet.Data is exhausted if only exhausted Packets
		// should be sent.
		if onlyFull && len(packet.Data) != int(packet.Header.Length) {
			// First not exhausted Packet found, return
			return nil
		}

		// TODO maybe check if data is empty - could be an issue

		n, err := packet.WriteTo(tdsChan.tdsConn.conn)
		if err != nil {
			return fmt.Errorf("error writing packet to server: %w", err)
		}

		if int(n) != int(packet.Header.Length)+MsgHeaderLength {
			return fmt.Errorf("expected to write %d bytes for packet, wrote %d instead",
				int(packet.Header.Length)+MsgHeaderLength, n)
		}
	}

	return nil
}

// WritePacket received packets from the associated TDSConn and attempts
// to produce Packages from the existing data.
func (tdsChan *TDSChannel) WritePacket(packet *Packet) {
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

	tdsChan.packageCh <- pkg
	return true
}
