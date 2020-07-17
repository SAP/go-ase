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
		queue, expectPacketQueue PacketQueue
	}{
		"no action": {
			queue: PacketQueue{
				queue: []*Packet{},
			},
			expectPacketQueue: PacketQueue{
				queue: []*Packet{},
			},
		},
		"shift by one by indexPacket": {
			queue: PacketQueue{
				queue: []*Packet{
					&Packet{},
				},
				indexPacket: 1,
			},
			expectPacketQueue: PacketQueue{
				queue: []*Packet{},
			},
		},
		"shift by one by indexData": {
			queue: PacketQueue{
				queue: []*Packet{
					&Packet{Data: make([]byte, 35)},
				},
				indexData: 35,
			},
			expectPacketQueue: PacketQueue{
				queue: []*Packet{},
			},
		},
		"shift multiple by both": {
			queue: PacketQueue{
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
			expectPacketQueue: PacketQueue{
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

	for title, cas := range cases {
		t.Run(title,
			func(t *testing.T) {
				cas.queue.DiscardUntilCurrentPosition()

				if !reflect.DeepEqual(cas.expectPacketQueue, cas.queue) {
					t.Errorf("Invalid result:")
					t.Errorf("Expected: %#v", cas.expectPacketQueue)
					t.Errorf("Received: %#v", cas.queue)
				}
			},
		)
	}
}

func TestPacketQueue_Bytes(t *testing.T) {
	cases := map[string]struct {
		queue, expectPacketQueue PacketQueue
		readBytes                int
		expectBytes              []byte
		expectErr                error
	}{
		"no action": {
			queue:       PacketQueue{},
			readBytes:   0,
			expectBytes: []byte{},
			expectErr:   nil,
		},
		"read byte": {
			queue: PacketQueue{
				queue: []*Packet{
					&Packet{Data: []byte{0x1}},
				},
			},
			readBytes:   1,
			expectBytes: []byte{0x1},
			expectErr:   nil,
		},
		"read over multiple queue": {
			queue: PacketQueue{
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

	for title, cas := range cases {
		t.Run(title,
			func(t *testing.T) {
				recvBytes, recvErr := cas.queue.Bytes(cas.readBytes)

				if cas.expectErr != nil {
					if recvErr != nil {
						t.Errorf("Did not receive expected error: %v", cas.expectErr)
						return
					}

					if cas.expectErr != recvErr {
						t.Errorf("Received unexpected error:")
						t.Errorf("Expected: %v", cas.expectErr)
						t.Errorf("Received: %v", recvErr)
						return
					}
					// An error was expected, return here
					return
				}

				if !bytes.Equal(cas.expectBytes, recvBytes) {
					t.Errorf("Received unexpected bytes:")
					t.Errorf("Expected: %#v", cas.expectBytes)
					t.Errorf("Received: %#v", recvBytes)
					return
				}
			},
		)
	}
}

func TestPacketQueue_WriteBytes(t *testing.T) {
	cases := map[string]struct {
		queue, expectPacketQueue PacketQueue
		writeBytes               [][]byte
	}{
		"no action": {
			queue:             PacketQueue{},
			expectPacketQueue: PacketQueue{},
			writeBytes:        [][]byte{[]byte{}},
		},
		"write byte": {
			queue: PacketQueue{},
			expectPacketQueue: PacketQueue{
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

	for title, cas := range cases {
		t.Run(title,
			func(t *testing.T) {
				for _, bs := range cas.writeBytes {
					err := cas.queue.WriteBytes(bs)
					if err != nil {
						t.Errorf("Received unexpected error: %v", err)
						return
					}
				}

				if !reflect.DeepEqual(cas.expectPacketQueue, cas.queue) {
					t.Errorf("Invalid result:")
					t.Errorf("Expected: %#v", cas.expectPacketQueue)
					t.Errorf("Received: %#v", cas.queue)
				}
			},
		)
	}
}
