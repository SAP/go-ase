// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"bytes"
	"reflect"
	"testing"
)

// fakePacket is a utility function to produce a valid Packet with
// pre-filled data.
func fakePacket(bs ...byte) *Packet {
	packet := NewPacket()
	copy(packet.Data, bs)
	return packet
}

func TestPacketQueue_DiscardUntilCurrentPosition(t *testing.T) {
	cases := map[string]struct {
		queue, expectPacketQueue *PacketQueue
	}{
		"no action": {
			queue: &PacketQueue{
				queue: []*Packet{},
			},
			expectPacketQueue: &PacketQueue{
				queue: []*Packet{},
			},
		},
		"shift by one by indexPacket": {
			queue: &PacketQueue{
				queue: []*Packet{
					&Packet{},
				},
				indexPacket: 1,
			},
			expectPacketQueue: &PacketQueue{
				queue: []*Packet{},
			},
		},
		"shift by one by indexData": {
			queue: &PacketQueue{
				queue: []*Packet{
					&Packet{Data: make([]byte, 35)},
				},
				indexData: 35,
			},
			expectPacketQueue: &PacketQueue{
				queue: []*Packet{},
			},
		},
		"shift multiple by both": {
			queue: &PacketQueue{
				queue: []*Packet{
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
				},
				indexPacket: 3,
				indexData:   14,
			},
			expectPacketQueue: &PacketQueue{
				queue: []*Packet{
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
					&Packet{Data: make([]byte, 35)},
				},
				indexPacket: 0,
				indexData:   14,
			},
		},
	}

	for title := range cases {
		t.Run(title,
			func(t *testing.T) {
				cases[title].queue.DiscardUntilCurrentPosition()

				if !reflect.DeepEqual(cases[title].expectPacketQueue, cases[title].queue) {
					t.Errorf("Invalid result:")
					t.Errorf("Expected: %#v", cases[title].expectPacketQueue)
					t.Errorf("Received: %#v", cases[title].queue)
				}
			},
		)
	}
}

func TestPacketQueue_Bytes(t *testing.T) {
	cases := map[string]struct {
		queue, expectPacketQueue *PacketQueue
		readBytes                int
		expectBytes              []byte
		expectErr                error
	}{
		"no action": {
			queue:       &PacketQueue{},
			readBytes:   0,
			expectBytes: []byte{},
			expectErr:   nil,
		},
		"read byte": {
			queue: &PacketQueue{
				queue: []*Packet{
					&Packet{Data: []byte{0x1}},
				},
			},
			readBytes:   1,
			expectBytes: []byte{0x1},
			expectErr:   nil,
		},
		"read over multiple queue": {
			queue: &PacketQueue{
				queue: []*Packet{
					&Packet{Data: []byte{0x1, 0x2}},
					&Packet{Data: []byte{0x3}},
					&Packet{Data: []byte{0x4, 0x5}},
				},
			},
			readBytes:   5,
			expectBytes: []byte{0x1, 0x2, 0x3, 0x4, 0x5},
			expectErr:   nil,
		},
	}

	for title := range cases {
		t.Run(title,
			func(t *testing.T) {
				recvBytes, recvErr := cases[title].queue.Bytes(cases[title].readBytes)

				if cases[title].expectErr != nil {
					if recvErr != nil {
						t.Errorf("Did not receive expected error: %v", cases[title].expectErr)
						return
					}

					if cases[title].expectErr != recvErr {
						t.Errorf("Received unexpected error:")
						t.Errorf("Expected: %v", cases[title].expectErr)
						t.Errorf("Received: %v", recvErr)
						return
					}
					// An error was expected, return here
					return
				}

				if !bytes.Equal(cases[title].expectBytes, recvBytes) {
					t.Errorf("Received unexpected bytes:")
					t.Errorf("Expected: %#v", cases[title].expectBytes)
					t.Errorf("Received: %#v", recvBytes)
					return
				}
			},
		)
	}
}

func TestPacketQueue_WriteBytes(t *testing.T) {
	cases := map[string]struct {
		queue, expectPacketQueue *PacketQueue
		writeBytes               [][]byte
	}{
		"no action": {
			queue:             &PacketQueue{},
			expectPacketQueue: &PacketQueue{},
			writeBytes:        [][]byte{[]byte{}},
		},
		"write byte": {
			queue: &PacketQueue{},
			expectPacketQueue: &PacketQueue{
				queue: []*Packet{
					fakePacket(0x1),
				},
				indexData: 1,
			},
			writeBytes: [][]byte{
				[]byte{0x1},
			},
		},
	}

	for title := range cases {
		t.Run(title,
			func(t *testing.T) {
				for _, bs := range cases[title].writeBytes {
					err := cases[title].queue.WriteBytes(bs)
					if err != nil {
						t.Errorf("Received unexpected error: %v", err)
						return
					}
				}

				if !reflect.DeepEqual(cases[title].expectPacketQueue, cases[title].queue) {
					t.Errorf("Invalid result:")
					t.Errorf("Expected: %#v", cases[title].expectPacketQueue)
					t.Errorf("Received: %#v", cases[title].queue)
				}
			},
		)
	}
}
