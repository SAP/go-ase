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
	packet := NewPacket(fakePacketSize())
	copy(packet.Data, bs)
	return packet
}

// fakePacketSize is a utility function to return a fake packet size
func fakePacketSize() int {
	return 512
}

// prepQueue returns a prepared PacketQueue for testing.
func prepQueue(idxPacket, idxData int, packets ...*Packet) *PacketQueue {
	for i, packet := range packets {
		if packet == nil {
			packets[i] = fakePacket()
		}
	}

	return &PacketQueue{
		queue:       packets,
		indexPacket: idxPacket,
		indexData:   idxData,
		packetSize:  fakePacketSize,
	}
}

// packetQueueEqual deeply compares two packet queues.
func packetQueueEqual(a, b *PacketQueue) bool {
	if a.indexPacket != b.indexPacket {
		return false
	}

	if a.indexData != b.indexData {
		return false
	}

	if a.packetSize() != b.packetSize() {
		return false
	}

	if len(a.queue) != len(b.queue) {
		return false
	}

	for i := range a.queue {
		if len(a.queue[i].Data) != len(b.queue[i].Data) {
			return false
		}

		if !reflect.DeepEqual(a.queue[i].Data, b.queue[i].Data) {
			return false
		}
	}

	return true
}

func TestPacketQueue_DiscardUntilCurrentPosition(t *testing.T) {
	cases := map[string]struct {
		queue, expectPacketQueue *PacketQueue
	}{
		"no action": {
			queue:             prepQueue(0, 0),
			expectPacketQueue: prepQueue(0, 0),
		},
		"shift by one by indexPacket": {
			queue:             prepQueue(1, 0, nil, nil),
			expectPacketQueue: prepQueue(0, 0, nil),
		},
		"shift by one by indexData": {
			queue:             prepQueue(0, fakePacketSize(), nil),
			expectPacketQueue: prepQueue(0, 0),
		},
		"shift multiple by both": {
			queue: prepQueue(3, 14,
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
			),
			expectPacketQueue: prepQueue(0, 14,
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
				&Packet{Data: make([]byte, 35)},
			),
		},
	}

	for title := range cases {
		t.Run(title,
			func(t *testing.T) {
				cases[title].queue.DiscardUntilCurrentPosition()

				if !packetQueueEqual(cases[title].expectPacketQueue, cases[title].queue) {
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
			queue:       prepQueue(0, 0),
			readBytes:   0,
			expectBytes: []byte{},
			expectErr:   nil,
		},
		"read byte": {
			queue:       prepQueue(0, 0, &Packet{Data: []byte{0x1}}),
			readBytes:   1,
			expectBytes: []byte{0x1},
			expectErr:   nil,
		},
		"read over multiple queue": {
			queue: prepQueue(0, 0,
				&Packet{Data: []byte{0x1, 0x2}},
				&Packet{Data: []byte{0x3}},
				&Packet{Data: []byte{0x4, 0x5}},
			),
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
			queue:             prepQueue(0, 0),
			expectPacketQueue: prepQueue(0, 0),
			writeBytes:        [][]byte{[]byte{}},
		},
		"write byte": {
			queue:             prepQueue(0, 0),
			expectPacketQueue: prepQueue(0, 1, fakePacket(0x1)),
			writeBytes: [][]byte{
				[]byte{0x1},
			},
		},
	}

	for title := range cases {
		t.Run(title,
			func(t *testing.T) {
				for _, bs := range cases[title].writeBytes {
					if err := cases[title].queue.WriteBytes(bs); err != nil {
						t.Errorf("Received unexpected error: %v", err)
						return
					}
				}

				if !packetQueueEqual(cases[title].expectPacketQueue, cases[title].queue) {
					t.Errorf("Invalid result:")
					t.Errorf("Expected: %#v", cases[title].expectPacketQueue)
					t.Errorf("Received: %#v", cases[title].queue)
				}
			},
		)
	}
}
