package tds

import (
	"errors"
	"fmt"
	"io"
)

var _ BytesChannel = (*PacketQueue)(nil)

type PacketQueue struct {
	queue                  []*Packet
	indexPacket, indexData int
}

func NewPacketQueue() *PacketQueue {
	return &PacketQueue{
		queue: []*Packet{},
	}
}

func (queue *PacketQueue) AddPacket(packet *Packet) {
	queue.queue = append(queue.queue, packet)
}

func (queue PacketQueue) Position() (int, int) {
	return queue.indexPacket, queue.indexData
}

func (queue *PacketQueue) SetPosition(indexPacket, indexData int) {
	queue.indexPacket = indexPacket
	queue.indexData = indexData
}

func (queue *PacketQueue) DiscardUntilCurrentPosition() {
	// .indexPacket points to no particular packet or no queue have
	// been consumed, empty stored queue and reset indizes.
	if len(queue.queue) == 0 || queue.indexPacket >= len(queue.queue) {
		queue.queue = []*Packet{}
		queue.indexPacket = 0
		queue.indexData = 0
		return
	}

	// shift queue
	queue.queue = queue.queue[queue.indexPacket:]
	queue.indexPacket = 0

	// If indexData is the end of the indexPacket the packet itself can
	// be discarded as well.
	if queue.indexData >= len(queue.queue[queue.indexPacket].Data) {
		queue.queue = queue.queue[1:]
		queue.indexData = 0
	}
}

// Read satisfies the io.Reader interface
func (queue *PacketQueue) Read(p []byte) (int, error) {
	var err error
	for i := range p {
		p[i], err = queue.Byte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return i, io.EOF
			}
			return i, err
		}
	}

	return len(p), nil
}

// Write satisfies the io.Writer interface
func (queue *PacketQueue) Write(p []byte) (int, error) {
	err := queue.WriteBytes(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Read methods

// Bytes returns at most n bytes as a slice.
//
// If the channel is closed before n bytes could be read Bytes will
// return a slice of length n with an io.EOF.
func (queue *PacketQueue) Bytes(n int) ([]byte, error) {
	if n == 0 {
		return []byte{}, nil
	}

	bs := make([]byte, n)
	// bsOffset is the index in the return slice where data still needs
	// to be written.
	bsOffset := 0

	for {
		if queue.indexPacket > len(queue.queue) {
			return nil, fmt.Errorf("not enough queue")
		}
		data := queue.queue[queue.indexPacket].Data

		startIndex := queue.indexData
		// (n - bsOffset) is the amount of bytes that still need to be
		// read.
		endIndex := queue.indexData + (n - bsOffset)
		if endIndex > len(data) {
			endIndex = len(data)
		}

		copy(bs[bsOffset:], data[startIndex:endIndex])
		bsOffset += endIndex - startIndex

		queue.indexData = endIndex
		// Move indizes forward if the current packet is consumed
		// entirely.
		if queue.indexData == len(data) {
			queue.indexPacket += 1
			queue.indexData = 0
		}

		if bsOffset == n {
			break
		}
	}

	return bs, nil
}

func (queue *PacketQueue) Byte() (byte, error) {
	bs, err := queue.Bytes(1)
	return bs[0], err
}

func (queue *PacketQueue) Uint8() (uint8, error) {
	b, err := queue.Byte()
	return uint8(b), err
}

func (queue *PacketQueue) Int8() (int8, error) {
	b, err := queue.Byte()
	return int8(b), err
}

func (queue *PacketQueue) Uint16() (uint16, error) {
	bs, err := queue.Bytes(2)
	return endian.Uint16(bs), err
}

func (queue *PacketQueue) Int16() (int16, error) {
	i, err := queue.Uint16()
	return int16(i), err
}

func (queue *PacketQueue) Uint32() (uint32, error) {
	bs, err := queue.Bytes(4)
	return endian.Uint32(bs), err
}

func (queue *PacketQueue) Int32() (int32, error) {
	i, err := queue.Uint32()
	return int32(i), err
}

func (queue *PacketQueue) Uint64() (uint64, error) {
	bs, err := queue.Bytes(8)
	return endian.Uint64(bs), err
}

func (queue *PacketQueue) Int64() (int64, error) {
	i, err := queue.Uint64()
	return int64(i), err
}

func (queue *PacketQueue) String(size int) (string, error) {
	bs, err := queue.Bytes(size)
	return string(bs), err
}

// Write methods

// WriteBytes writes a slice of bytes.
//
// An error is only returned if the channel is marked as closed when
// starting to pass bytes to the underlying channel.
func (queue *PacketQueue) WriteBytes(bs []byte) error {
	if len(bs) == 0 {
		return nil
	}

	bsOffset := 0

	for bsOffset < len(bs) {
		// Add new packet if the index points to no packet
		if queue.indexPacket == len(queue.queue) {
			queue.queue = append(queue.queue, NewPacket())
		}

		// Retrieve current package and calculate how many bytes can
		// still be written to it.
		curPacket := queue.queue[queue.indexPacket]
		freeBytes := int(curPacket.Header.Length) - queue.indexData

		// No free bytes, add a new packet.
		if freeBytes == 0 {
			curPacket = &Packet{}
			queue.queue = append(queue.queue, curPacket)
			queue.indexPacket++
			queue.indexData = 0
			freeBytes = int(curPacket.Header.Length)
		}

		// Calculate how many bytes are left in bs if more free bytes
		// are available in the packet than are left in bs.
		if freeBytes > len(bs)-bsOffset {
			freeBytes = len(bs) - bsOffset
		}

		// curPacket.Data = append(curPacket.Data, bs[bsOffset:bsOffset+freeBytes]...)
		copy(curPacket.Data[queue.indexData:], bs[bsOffset:bsOffset+freeBytes])
		bsOffset += freeBytes
		queue.indexData += freeBytes
	}

	return nil
}

func (queue *PacketQueue) WriteByte(b byte) error {
	return queue.WriteBytes([]byte{b})
}

func (queue *PacketQueue) WriteUint8(i uint8) error {
	return queue.WriteByte(byte(i))
}

func (queue *PacketQueue) WriteInt8(i int8) error {
	return queue.WriteUint8(uint8(i))
}

func (queue *PacketQueue) WriteUint16(i uint16) error {
	bs := make([]byte, 2)
	endian.PutUint16(bs, i)
	return queue.WriteBytes(bs)
}

func (queue *PacketQueue) WriteInt16(i int16) error {
	return queue.WriteUint16(uint16(i))
}

func (queue *PacketQueue) WriteUint32(i uint32) error {
	bs := make([]byte, 4)
	endian.PutUint32(bs, i)
	return queue.WriteBytes(bs)
}

func (queue *PacketQueue) WriteInt32(i int32) error {
	return queue.WriteUint32(uint32(i))
}

func (queue *PacketQueue) WriteUint64(i uint64) error {
	bs := make([]byte, 8)
	endian.PutUint64(bs, i)
	return queue.WriteBytes(bs)
}

func (queue *PacketQueue) WriteInt64(i int64) error {
	return queue.WriteUint64(uint64(i))
}

func (queue *PacketQueue) WriteString(s string) error {
	return queue.WriteBytes([]byte(s))
}
