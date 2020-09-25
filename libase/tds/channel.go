// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

var (
	ErrNoPackageReady = errors.New("no package ready")
	ErrChannelClosed  = errors.New("channel is closed")
)

// Channel is a channel in a multiplexed connection with a TDS
// server.
type Channel struct {
	tdsConn *Conn

	// The RWMutex isn't used in it intended from of reader/writer locks
	// but rather allows multiple goroutines to acquire a read-lock to
	// use the channel for sending/receiving packages.
	// The exclusive write lock is used to stop other goroutines from
	// using the channel when closing it.
	sync.RWMutex
	closed bool

	channelId int

	envChangeHooks     []EnvChangeHook
	envChangeHooksLock *sync.Mutex

	eedHooks     []EEDHook
	eedHooksLock *sync.Mutex

	// currentHeaderType is the PacketHeaderType set on outgoing
	// packets.
	CurrentHeaderType PacketHeaderType
	// curPacketNr is the number of the next packet being sent
	curPacketNr int
	// window is the amount of buffers transmitted between ACKs
	window int

	// queues store unconsumed Packets
	queueRx, queueTx *PacketQueue
	// lastPkgRx/Tx are the last packages sent to/received from the TDS
	// server
	lastPkgRx, lastPkgTx Package
	// packageCh stores Packages as they are parsed from Packets
	packageCh chan Package

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
		eedHooks:           []EEDHook{},
		eedHooksLock:       &sync.Mutex{},
		CurrentHeaderType:  TDS_BUF_NORMAL,
		window:             0, // TODO
		queueRx:            NewPacketQueue(tds.PacketSize),
		queueTx:            NewPacketQueue(tds.PacketSize),
		packageCh:          make(chan Package, queueSize),
		errCh:              make(chan error, 10),
	}

	tds.tdsChannels[channelId] = tdsChan

	// channel 0 needs no setup
	if channelId == 0 {
		return tdsChan, nil
	}

	// Send packets to setup logical channel
	setup := NewPacket(PacketHeaderSize)
	setup.Header.Length = PacketHeaderSize
	setup.Data = nil

	tdsChan.CurrentHeaderType = TDS_BUF_SETUP
	err = tdsChan.sendPacket(setup)
	if err != nil {
		return nil, fmt.Errorf("error sending setup for channel %d: %w",
			tdsChan.channelId, err)
	}

	pkg, err := tdsChan.NextPackage(context.Background(), true)
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
	tdsChan.RLock()
	defer tdsChan.RUnlock()
	if tdsChan.closed {
		return
	}

	tdsChan.CurrentHeaderType = TDS_BUF_NORMAL
	tdsChan.queueTx.Reset()
	tdsChan.lastPkgTx = nil
}

// Close communicates the closing of the channel with the TDS server.
//
// The teardown on client side is guaranteed, even if Close returns an
// error. An error is only returned if the communication with the server
// fails or if packages or error remained in the channels.
//
// If an error is returned it is a *multierror.Error with all errors.
func (tdsChan *Channel) Close() error {
	var me error

	if tdsChan.channelId == 0 {
		// Channel 0 is the main communication channel - send logout packages
		if err := tdsChan.Logout(); err != nil {
			me = multierror.Append(me, fmt.Errorf("error in logout sequence: %w", err))
		}
	} else {
		// Closing of logical channels must be communicated using
		// header-only packets

		// Send packet to tear down logical channel
		teardown := NewPacket(tdsChan.tdsConn.PacketSize())
		teardown.Data = nil
		tdsChan.CurrentHeaderType = TDS_BUF_CLOSE

		if err := tdsChan.sendPacket(teardown); err != nil {
			me = multierror.Append(me,
				fmt.Errorf("error sending teardown for channel %d: %w",
					tdsChan.channelId, err))
		}

		// TODO process ack packet
	}

	// Lock the channel and store the closed indicator.
	tdsChan.Lock()
	defer tdsChan.Unlock()

	tdsChan.closed = true

	// Channel closing has been communicated, remove channel from conn
	tdsChan.tdsConn.tdsChannelsLock.Lock()
	delete(tdsChan.tdsConn.tdsChannels, tdsChan.channelId)
	tdsChan.tdsConn.tdsChannelsLock.Unlock()

	close(tdsChan.packageCh)
	for {
		if pkg, ok := <-tdsChan.packageCh; ok {
			me = multierror.Append(me, fmt.Errorf("package still queued: %v", pkg))
		} else {
			break
		}
	}
	tdsChan.packageCh = nil

	close(tdsChan.errCh)
	for {
		if err, ok := <-tdsChan.errCh; ok {
			me = multierror.Append(me, fmt.Errorf("error still queued: %w", err))
		} else {
			break
		}
	}
	tdsChan.errCh = nil

	return me
}

// Logout implements the logout sequence.
func (tdsChan *Channel) Logout() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := tdsChan.SendPackage(ctx, &LogoutPackage{})
	if err != nil {
		return fmt.Errorf("error sending logout package: %w", err)
	}

	pkg, err := tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading logout response: %w", err)
	}

	done, ok := pkg.(*DonePackage)
	if !ok {
		return fmt.Errorf("expected done package in logout response, got: %v", pkg)
	}

	if done.Status&TDS_DONE_FINAL != TDS_DONE_FINAL {
		return fmt.Errorf("received done package with status %s instead of TDS_DONE_FINAL",
			done.Status)
	}

	if done.TranState&TDS_TRAN_COMPLETED != TDS_TRAN_COMPLETED {
		return fmt.Errorf("received done package with transtate %s instead of TDS_TRAN_COMPLETED",
			done.TranState)
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
			if member.Type == TDS_ENV_PACKSIZE {
				packSize, err := strconv.Atoi(member.NewValue)
				if err != nil {
					return false, fmt.Errorf("error parsing new packet size '%s' to int: %w",
						member.NewValue, err)
				}
				tdsChan.tdsConn.packetSize = packSize
			}

			tdsChan.callEnvChangeHooks(member.Type, member.OldValue, member.NewValue)
		}
		return false, nil
	}

	if eed, ok := pkg.(*EEDPackage); ok {
		// TDS_EED_INFO packages are not supposed to leave the client
		// library.
		if eed.Status&TDS_EED_INFO == TDS_EED_INFO {
			return false, nil
		}

		tdsChan.callEEDHooks(*eed)
		return true, nil
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
func (tdsChan *Channel) NextPackage(ctx context.Context, wait bool) (Package, error) {
	tdsChan.RLock()
	defer tdsChan.RUnlock()

	if tdsChan.closed {
		return nil, ErrChannelClosed
	}

	ch := make(chan error, 1)

	// Write an error into the channel if the caller does not want to
	// wait. The channel will be empty otherwise, block the select and
	// realise the wait.
	if !wait {
		ch <- ErrNoPackageReady
	}

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-ctx.Done():
		return nil, context.Canceled
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

// NextPackageUntil calls NextPackage until the passed function
// processPkg returns true.
//
// If processPkg returns true no further packages will be consumed, so
// the communication handling can be passed to another function.
//
// If processPkg returns an error all packages in the payload will be
// consumed and an error containing all EEDPackages in the payload will
// be returned.
//
// If processPkg is nil all packages in the current payload are consumed
// and no package and an io.EOF error is returned.
// The io.EOF is wrapped with an EEDError with all EEDPackages in the
// payload if the payload contained any EEDPackages.
//
// To just consume all packages a consumer can return an io.EOF in
// processPkg and check if the error is not of io.EOF:
//
// _, err := ...NextPackageUntil(ctx, wait, func(pkg tds.Package) (bool, error) {
//     switch pkg.(type) {
//     case ...
//         // handle communication
//     case ...
//         // handle final communication
//         return true, io.EOF
//     }
// }
// if err != nil && !errors.Is(err, io.EOF) {
//     // error handling
// }
func (tdsChan *Channel) NextPackageUntil(ctx context.Context, wait bool, processPkg func(Package) (bool, error)) (Package, error) {
	eedError := &EEDError{}

	for {
		pkg, err := tdsChan.NextPackage(ctx, wait)
		if err != nil {
			return nil, err
		}

		if eed, ok := pkg.(*EEDPackage); ok {
			eedError.Add(eed)
			continue
		}

		if processPkg == nil {
			if ok, _ := isDoneFinal(pkg); ok {
				return nil, io.EOF
			}

			_, err := tdsChan.NextPackageUntil(ctx, wait, func(pkg Package) (bool, error) {
				return isDoneFinal(pkg)
			})
			return nil, err
		}

		ok, err := processPkg(pkg)
		if err != nil {
			// Consume all packages until DonePackage{TDS_DONE_FINAL} if
			// the current package wasn't a DonePackage{TDS_DONE_FINAL}
			// to prevent any leftovers that may impact later
			// communications.
			if ok, _ := isDoneFinal(pkg); !ok {
				_, err := tdsChan.NextPackageUntil(ctx, wait, nil)
				// Append any additional received EEDPackages to the
				// EEDError.
				var finalEEDError *EEDError
				if err != nil && errors.As(err, &finalEEDError) {
					eedError.EEDPackages = append(eedError.EEDPackages, finalEEDError.EEDPackages...)
				}
			}

			err = fmt.Errorf("tds: error in user-defined processing function: %w", err)

			// Only return an EEDError if there were EEDPackages
			if len(eedError.EEDPackages) == 0 {
				return nil, err
			}

			eedError.WrappedError = err
			return nil, eedError
		}

		if ok {
			return pkg, nil
		}
	}
}

func isDoneFinal(pkg Package) (bool, error) {
	done, ok := pkg.(*DonePackage)
	return ok && done.Status == TDS_DONE_FINAL, nil
}

type LastPkgAcceptor interface {
	LastPkg(Package) error
}

// QueuePackage utilizes PacketQueue to convert a Package into packets.
// Packets that have their Data exhausted are sent to the server.
func (tdsChan *Channel) QueuePackage(ctx context.Context, pkg Package) error {
	tdsChan.RLock()
	defer tdsChan.RUnlock()
	// TODO return proper error
	if tdsChan.closed {
		return ErrChannelClosed
	}

	if acceptor, ok := pkg.(LastPkgAcceptor); ok {
		err := acceptor.LastPkg(tdsChan.lastPkgTx)
		if err != nil {
			return fmt.Errorf("error calling LastPkg on %s: %w", pkg, err)
		}
	}

	err := pkg.WriteTo(tdsChan.queueTx)
	if err != nil {
		return fmt.Errorf("error queueing packets from package %s: %w", pkg, err)
	}
	tdsChan.lastPkgTx = pkg

	return tdsChan.sendPackets(ctx, true)
}

// Send all remaining Packets in queue to the server.
// This includes Packets whose Data isn't exhausted.
func (tdsChan *Channel) SendRemainingPackets(ctx context.Context) error {
	tdsChan.RLock()
	defer tdsChan.RUnlock()
	if tdsChan.closed {
		return ErrChannelClosed
	}

	// SendRemainingPackets is only called when completing sending
	// packets to the server and preparing to receive the answer.
	defer tdsChan.Reset()
	return tdsChan.sendPackets(ctx, false)
}

// SendPackage combines calls to QueuePackage and SendRemainingPackets
// and can be used if e.g. the last package or only a single package
// must be sent.
func (tdsChan *Channel) SendPackage(ctx context.Context, pkg Package) error {
	if err := tdsChan.QueuePackage(ctx, pkg); err != nil {
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
			if i == tdsChan.queueTx.indexPacket && tdsChan.queueTx.indexData < tdsChan.tdsConn.PacketBodySize() {
				if onlyFull {
					// Packet is not exhausted and only exhausted packets
					// should be sent. Return.
					return nil
				}

				// Packet is not exhausted but should be sent. Adjust header
				// length
				packet.Header.Length = uint16(PacketHeaderSize + tdsChan.queueTx.indexData)
				packet.Data = packet.Data[:tdsChan.queueTx.indexData]
			}

			// TODO maybe check if data is empty - could be an issue

			err := tdsChan.sendPacket(packet)
			if err != nil {
				return fmt.Errorf("error sending packet %s: %w", packet, err)
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

	if len(packet.Data) != tdsChan.tdsConn.PacketBodySize() {
		// Data portion is not exhausted, this is the last packet.
		packet.Header.Status |= TDS_BUFSTAT_EOM
	}

	n, err := packet.WriteTo(tdsChan.tdsConn.conn)
	if err != nil {
		return fmt.Errorf("error writing packet to server: %w", err)
	}

	if int(n) != int(packet.Header.Length) {
		return fmt.Errorf("expected to write %d bytes for packet, wrote %d instead",
			int(packet.Header.Length)+PacketHeaderSize, n)
	}

	return nil
}

// WritePacket received packets from the associated Conn and attempts
// to produce Packages from the existing data.
func (tdsChan *Channel) WritePacket(packet *Packet) {
	tdsChan.RLock()
	defer tdsChan.RUnlock()
	if tdsChan.closed {
		return
	}

	// The packet is header-only - pass it directly into the package
	// channel.
	if packet.Header.Length == PacketHeaderSize {
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
		if ok := tdsChan.tryParsePackage(); !ok {
			// Attempting to parse package failed
			if tdsChan.queueRx.IsEOM() {
				// And queue is EOM - reset queue
				tdsChan.queueRx.Reset()
			} else {
				// Roll back position and return.
				tdsChan.queueRx.SetPosition(curPacket, curData)
			}
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
		if tdsChan.queueRx.IsEOM() {
			// If the error is io.EOF then the payload from the server
			// has been fully consumed.
			// TDS doesn't always send a DonePackage with TDS_DONE_FINAL
			// - usually only when a procedure with multiple commands is
			// being executed.
			if lastPkg, ok := tdsChan.lastPkgRx.(*DonePackage); !ok || lastPkg.Status != TDS_DONE_FINAL {
				tdsChan.packageCh <- &DonePackage{Status: TDS_DONE_FINAL}
			}
		}
		return false
	}

	// Create Package.
	pkg, err := LookupPackage(Token(tokenByte))
	if err != nil {
		tdsChan.errCh <- err
		return false
	}

	// If the Package is tokenless write the token byte back in.
	if tokenless, ok := pkg.(*TokenlessPackage); ok {
		tokenless.Data.WriteByte(tokenByte)
	}

	if acceptor, ok := pkg.(LastPkgAcceptor); ok {
		if err := acceptor.LastPkg(tdsChan.lastPkgRx); err != nil {
			tdsChan.errCh <- fmt.Errorf("error in LastPkg: %w", err)
			return false
		}
	}

	// Read data into Package.
	if err := pkg.ReadFrom(tdsChan.queueRx); err != nil {
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
	tdsChan.lastPkgRx = pkg
	return true
}
