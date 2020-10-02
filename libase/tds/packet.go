// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrEOFAfterZeroRead = errors.New("received io.EOF after reading 0 bytes")
)

// Packet represents a single packet in a message.
type Packet struct {
	Header PacketHeader
	Data   []byte
}

func NewPacket(packetSize int) *Packet {
	packet := &Packet{}
	packet.Header = NewPacketHeader(packetSize)
	packet.Data = make([]byte, packetSize-PacketHeaderSize)
	return packet
}

func (packet Packet) Bytes() ([]byte, error) {
	bs := make([]byte, int(packet.Header.Length))

	_, err := packet.Header.Read(bs[:PacketHeaderSize])
	if err != nil {
		return nil, fmt.Errorf("error reading header into byte slice: %w", err)
	}

	copy(bs[PacketHeaderSize:], packet.Data)
	return bs, nil
}

func (packet *Packet) ReadFrom(ctx context.Context, reader io.Reader, timeout time.Duration) (int64, error) {
	var totalBytes int64

	packet.Header = PacketHeader{}
	n, err := packet.Header.ReadFrom(reader)
	if err != nil {
		return n, fmt.Errorf("failed to read header: %w", err)
	}

	totalBytes += n

	packet.Data = make([]byte, packet.Header.Length-PacketHeaderSize)

	// The timeout will be refreshed (replaced) on every successful
	// read. This is done so the timeout only triggers if there was
	// actually no data read from the server to prevent failures when
	// the PDU is split over multiple responses and the responses
	// themselves are arriving slowly due to the network (e.g. erroneous
	// scheduling, packet inspection, overloaded firewalls, etc.pp.).
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		if err := ctx.Err(); err != nil {
			return totalBytes, err
		}

		m, err := reader.Read(packet.Data[totalBytes-n:])
		totalBytes += int64(m)

		if m > 0 {
			timeoutCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				// Check if the timeout was exceeded _and_ if the last
				// read returned 0 bytes. So the timeout may be
				// exceeded but isn't triggered until the packet also
				// can't read any more data.
				if err := timeoutCtx.Err(); err != nil && m == 0 {
					return totalBytes, ErrEOFAfterZeroRead
				}

				// The PDU is split over multiple responses
				if totalBytes != int64(packet.Header.Length) {
					continue
				}

				if packet.Header.MsgType == TDS_BUF_CLOSE {
					return totalBytes, err
				}
			}

			return totalBytes, fmt.Errorf("error reading body: %w", err)
		}

		if totalBytes == int64(packet.Header.Length) {
			// Read the expected amount of bytes
			break
		}
	}

	return totalBytes, nil
}

func (packet Packet) WriteTo(writer io.Writer) (int64, error) {
	bs, err := packet.Bytes()
	if err != nil {
		return 0, fmt.Errorf("error compiling packet bytes: %w", err)
	}

	n, err := writer.Write(bs)
	return int64(n), err
}

func (packet Packet) String() string {
	strHeaderStatus := deBitmaskString(int(packet.Header.Status), int(TDS_BUFSTAT_SYMENCRYPT),
		func(i int) string { return PacketHeaderStatus(i).String() },
		"no status",
	)

	return fmt.Sprintf(
		"Type: %s, Status: %s, Length: %d, Channel: %d, PacketNr: %d, Window: %d, DataLen: %d",
		packet.Header.MsgType,
		strHeaderStatus,
		packet.Header.Length,
		packet.Header.Channel,
		packet.Header.PacketNr,
		packet.Header.Window,
		len(packet.Data),
	)
}
