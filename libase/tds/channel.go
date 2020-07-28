package tds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/hashicorp/go-multierror"
)

var ErrNoPackageReady = errors.New("no package ready")

// Channel is a channel in a multiplexed connection with a TDS
// server.
type Channel struct {
	tdsConn *Conn

	channelId          int
	envChangeHooks     []EnvChangeHook
	envChangeHooksLock *sync.Mutex

	// currentHeaderType is the PacketHeaderType set on outgoing
	// packets.
	CurrentHeaderType PacketHeaderType
	// curPacketNr is the number of the next packet being sent
	curPacketNr int
	// window is the amount of buffers transmitted between ACKs
	window int

	// queues store unconsumed Packets
	queueRx, queueTx *PacketQueue
	// packageCh stores Packages as they are parsed from Packets
	packageCh chan Package
	// lastPkg is the last package sent by the TDS server and added to
	// the channel.
	lastPkg Package

	errCh chan error
}

// NewChannel communicates the creation of a new channel with the
// server.
func (tds *Conn) NewChannel() (*Channel, error) {
	channelId, err := tds.getValidChannelId()
	if err != nil {
		return nil, fmt.Errorf("error getting channel ID: %w", err)
	}

	queueSize := 100
	if prop := tds.dsn.Prop("channel-package-queue-size"); prop != "" {
		queueSize, err = strconv.Atoi(prop)
		if err != nil {
			return nil, fmt.Errorf("error parsing number from channel-package-queue-size '%s': %w", prop, err)
		}
	}

	tdsChan := &Channel{
		tdsConn:            tds,
		channelId:          channelId,
		envChangeHooks:     []EnvChangeHook{},
		envChangeHooksLock: &sync.Mutex{},
		CurrentHeaderType:  TDS_BUF_NORMAL,
		window:             0, // TODO
		queueRx:            NewPacketQueue(),
		queueTx:            NewPacketQueue(),
		packageCh:          make(chan Package, queueSize),
		errCh:              make(chan error, 10),
	}

	tds.tdsChannels[channelId] = tdsChan

	// channel 0 needs no setup
	if channelId == 0 {
		return tdsChan, nil
	}

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

	tdsChan.Reset()
	return tdsChan, nil
}

// Reset resets the Channel after a communication has been completed.
func (tdsChan *Channel) Reset() {
	tdsChan.CurrentHeaderType = TDS_BUF_NORMAL
	tdsChan.queueTx.Reset()
	tdsChan.lastPkg = nil
}

// Close communicates the closing of the channel with the TDS server.
//
// The teardown on client side is guaranteed, even if Close returns an
// error. An error is only returned if the communication with the server
// fails or if packages or error remained in the channels.
//
// If an error is returned it is a *multierror.Error with all errors.
func (tdsChan *Channel) Close() error {
	// Remove channel from connection
	delete(tdsChan.tdsConn.tdsChannels, tdsChan.channelId)

	var me error

	if tdsChan.channelId == 0 {
		// Channel 0 is the main communication channel - send logout packages
		err := tdsChan.Logout()
		if err != nil {
			me = multierror.Append(me, fmt.Errorf("error in logout sequence: %w", err))
		}
	} else {
		// Closing of logical channels must be communicated using
		// header-only packets

		// Send packet to tear down logical channel
		teardown := NewPacket()
		teardown.Header.Length = MsgHeaderLength
		teardown.Data = nil
		tdsChan.CurrentHeaderType = TDS_BUF_CLOSE

		err := tdsChan.sendPacket(teardown)
		if err != nil {
			me = multierror.Append(me,
				fmt.Errorf("error sending teardown for channel %d: %w",
					tdsChan.channelId, err))
		}
	}

	close(tdsChan.packageCh)
	for {
		if pkg, ok := <-tdsChan.packageCh; ok {
			me = multierror.Append(me, fmt.Errorf("package still queued: %v", pkg))
		} else {
			break
		}
	}

	close(tdsChan.errCh)
	for {
		if err, ok := <-tdsChan.errCh; ok {
			me = multierror.Append(me, fmt.Errorf("error still queued: %w", err))
		} else {
			break
		}
	}

	return me
}

// Logout implements the logout sequence.
func (tdsChan Channel) Logout() error {
	err := tdsChan.SendPackage(context.Background(), &LogoutPackage{})
	if err != nil {
		return fmt.Errorf("error sending logout package: %w", err)
	}

	pkg, err := tdsChan.NextPackage(true)
	if err != nil {
		return fmt.Errorf("error reading logout response: %w", err)
	}

	done, ok := pkg.(*DonePackage)
	if !ok {
		return fmt.Errorf("expected done package in logout response, got: %v", pkg)
	}

	if done.status|TDS_DONE_FINAL != TDS_DONE_FINAL {
		return fmt.Errorf("received done package with status %s instead of TDS_DONE_FINAL",
			done.status)
	}

	if done.tranState|TDS_TRAN_COMPLETED != TDS_TRAN_COMPLETED {
		return fmt.Errorf("received done package with transtate %s instead of TDS_TRAN_COMPLETED",
			done.tranState)
	}

	return nil
}

// handleSpecialPackage handles special packages such as env changes.
// The returned boolean signals if the package should be passed along or
// skipped.
// An error is returned if the handling errored.
func (tdsChan *Channel) handleSpecialPackage(pkg Package) (bool, error) {
	if envChange, ok := pkg.(*EnvChangePackage); ok {
		for _, member := range envChange.members {
			tdsChan.callEnvChangeHooks(member.Type, member.NewValue, member.OldValue)
		}
		return false, nil
	}

	return true, nil
}

// NextPackage returns the next package in the queue.
// An error may be returned in the following cases:
//	1. The connections' context was closed.
//	2. The connection has a communication error queued.
//	3. The channel has a parsing error queued.
//
// If wait is false a ErrNoPackageReady error may be returned.
//
// If multiple errors and a package are ready a random error or package
// will be returned, as stated in the spec for select.
func (tdsChan *Channel) NextPackage(wait bool) (Package, error) {
	ch := make(chan error, 1)

	// Write an error into the channel if the caller does not want to
	// wait. The channel will be empty otherwise, block the select and
	// realise the wait.
	if !wait {
		ch <- ErrNoPackageReady
	}

	select {
	case <-tdsChan.tdsConn.ctx.Done():
		return nil, context.Canceled
	case err := <-tdsChan.tdsConn.errCh:
		return nil, fmt.Errorf("error in TDS connection: %w", err)
	case err := <-tdsChan.errCh:
		return nil, fmt.Errorf("error in TDS channel %d: %w",
			tdsChan.channelId, err)
	case pkg := <-tdsChan.packageCh:
		return pkg, nil
	case err := <-ch:
		return nil, err
	}
}

type LastPkgAcceptor interface {
	LastPkg(Package) error
}

// QueuePackage utilizes PacketQueue to convert a Package into packets.
// Packets that have their Data exhausted are sent to the server.
func (tdsChan *Channel) QueuePackage(ctx context.Context, pkg Package) error {
	if acceptor, ok := pkg.(LastPkgAcceptor); ok {
		err := acceptor.LastPkg(tdsChan.lastPkg)
		if err != nil {
			return fmt.Errorf("error calling LastPkg: %w", err)
		}
	}

	err := pkg.WriteTo(tdsChan.queueTx)
	if err != nil {
		return fmt.Errorf("error queueing packets from package: %w", err)
	}
	tdsChan.lastPkg = pkg

	return tdsChan.sendPackets(ctx, true)
}

// Send all remaining Packets in queue to the server.
// This includes Packets whose Data isn't exhausted.
func (tdsChan *Channel) SendRemainingPackets(ctx context.Context) error {
	// SendRemainingPackets is only called when completing sending
	// packets to the server and preparing to receive the answer.
	defer tdsChan.Reset()
	return tdsChan.sendPackets(ctx, false)
}

// SendPackage combines calls to QueuePackage and SendRemainingPackets
// and can be used if e.g. the last package or only a single package
// must be sent.
func (tdsChan *Channel) SendPackage(ctx context.Context, pkg Package) error {
	err := tdsChan.QueuePackage(ctx, pkg)
	if err != nil {
		return err
	}

	return tdsChan.SendRemainingPackets(ctx)
}

func (tdsChan *Channel) sendPackets(ctx context.Context, onlyFull bool) error {
	defer tdsChan.queueTx.DiscardUntilCurrentPosition()

	for i, packet := range tdsChan.queueTx.queue {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-tdsChan.tdsConn.ctx.Done():
			return context.Canceled
		default:
			// Only the last packet should not be full.
			if i == tdsChan.queueTx.indexPacket && tdsChan.queueTx.indexData < MsgBodyLength {
				if onlyFull {
					// Packet is not exhausted and only exhausted packets
					// should be sent. Return.
					return nil
				}

				// Packet is not exhausted but should be sent. Adjust header
				// length
				packet.Header.Length = uint16(MsgHeaderLength + tdsChan.queueTx.indexData)
				packet.Data = packet.Data[:tdsChan.queueTx.indexData]
			}

			// TODO maybe check if data is empty - could be an issue

			err := tdsChan.sendPacket(packet)
			if err != nil {
				return fmt.Errorf("error sending packet: %w", err)
			}
		}
	}

	return nil
}

func (tdsChan *Channel) sendPacket(packet *Packet) error {
	packet.Header.MsgType = tdsChan.CurrentHeaderType

	// Channel 0 does not need PacketNr or Window
	if tdsChan.channelId > 0 {
		packet.Header.Channel = uint16(tdsChan.channelId)
		packet.Header.PacketNr = uint8(tdsChan.curPacketNr)
		tdsChan.curPacketNr = (tdsChan.curPacketNr + 1) % 256
		packet.Header.Window = uint8(tdsChan.window)
	}

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

// WritePacket received packets from the associated Conn and attempts
// to produce Packages from the existing data.
func (tdsChan *Channel) WritePacket(packet *Packet) {
	// The packet is header-only - pass it directly into the package
	// channel.
	if packet.Header.Length == MsgHeaderLength {
		tdsChan.packageCh <- HeaderOnlyPackage{Header: packet.Header}
		return
	}

	// Add packet into queue
	tdsChan.queueRx.AddPacket(packet)

	for {
		// Read out current position for resetting if the existing data
		// isn't enough to fill a Package.
		curPacket, curData := tdsChan.queueRx.Position()

		// Attempt to parse a Package.
		ok := tdsChan.tryParsePackage()
		if !ok {
			// Attempt failed, roll back position and return.
			tdsChan.queueRx.SetPosition(curPacket, curData)
			return
		}

		// Package could be filled with the available data. Discard all
		// consumed packets.
		tdsChan.queueRx.DiscardUntilCurrentPosition()
	}
}

// tryParsePackage attempts to parse a Package from the queued Packets.
func (tdsChan *Channel) tryParsePackage() bool {
	// Attempt to process data from channel into a Package.
	tokenByte, err := tdsChan.queueRx.Byte()
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

	if acceptor, ok := pkg.(LastPkgAcceptor); ok {
		err := acceptor.LastPkg(tdsChan.lastPkg)
		if err != nil {
			tdsChan.errCh <- fmt.Errorf("error in LastPkg: %w", err)
			return false
		}
	}

	// Read data into Package.
	err = pkg.ReadFrom(tdsChan.queueRx)
	if err != nil {
		if errors.Is(err, ErrNotEnoughBytes) {
			// Not enough bytes available to parse package
			return false
		}

		// Parsing went wrong, record as error
		tdsChan.errCh <- fmt.Errorf("error parsing package: %w", err)
		return false
	}

	pass, err := tdsChan.handleSpecialPackage(pkg)
	if err != nil {
		tdsChan.errCh <- fmt.Errorf("error while handling special package: %w", err)
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
