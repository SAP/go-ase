package tds

import (
	"errors"
	"io"
)

// channel is used to abstract the conversion of data types away from
// the package read/write methods.
type channel struct {
	ch     chan byte
	closed bool
}

func newChannel() *channel {
	return &channel{
		// TODO channel size configurable? Without the complexity to
		// enable waiting until the channel has been exhausted the
		// channel should be able to carry roughly two packets.
		// If the channel can carry less than one packet it may result
		// in a deadlock where one package finished reading a packet
		// while the next packet is being written to the channel.
		// Since the packet is being written to the channel in the main
		// thread, which also swaps out the packages the channel can't
		// be cleared enough to allow the main thread to proceed.
		ch: make(chan byte, 1024),
	}
}

func (ch *channel) Close() {
	ch.closed = true
	close(ch.ch)
}

// Read satisfies the io.Reader interface
func (ch *channel) Read(p []byte) (int, error) {
	var err error
	for i := range p {
		p[i], err = ch.Byte()
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
func (ch *channel) Write(p []byte) (int, error) {
	err := ch.WriteBytes(p)
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
func (ch *channel) Bytes(n int) ([]byte, error) {
	if n == 0 {
		return []byte{}, nil
	}

	bs := make([]byte, n)
	for n := range bs {
		var ok bool
		// Keep reading until the channel provided a byte or until the
		// channel is closed.
		bs[n], ok = <-ch.ch
		// No byte could be read and the channel is marked as closed
		if !ok && ch.closed {
			return bs, io.EOF
		}
	}

	return bs, nil
}

func (ch *channel) Byte() (byte, error) {
	bs, err := ch.Bytes(1)
	return bs[0], err
}

func (ch *channel) Uint8() (uint8, error) {
	b, err := ch.Byte()
	return uint8(b), err
}

func (ch *channel) Int8() (int8, error) {
	b, err := ch.Byte()
	return int8(b), err
}

func (ch *channel) Uint16() (uint16, error) {
	bs, err := ch.Bytes(2)
	return endian.Uint16(bs), err
}

func (ch *channel) Int16() (int16, error) {
	i, err := ch.Uint16()
	return int16(i), err
}

func (ch *channel) Uint32() (uint32, error) {
	bs, err := ch.Bytes(4)
	return endian.Uint32(bs), err
}

func (ch *channel) Int32() (int32, error) {
	i, err := ch.Uint32()
	return int32(i), err
}

func (ch *channel) Uint64() (uint64, error) {
	bs, err := ch.Bytes(8)
	return endian.Uint64(bs), err
}

func (ch *channel) Int64() (int64, error) {
	i, err := ch.Uint64()
	return int64(i), err
}

func (ch *channel) String(size int) (string, error) {
	bs, err := ch.Bytes(size)
	return string(bs), err
}

// Write methods

// WriteBytes writes a slice of bytes.
//
// An error is only returned if the channel is marked as closed when
// starting to pass bytes to the underlying channel.
func (ch *channel) WriteBytes(bs []byte) error {
	if ch.closed {
		return io.EOF
	}

	for _, b := range bs {
		ch.ch <- b
	}

	return nil
}

func (ch *channel) WriteByte(b byte) error {
	return ch.WriteBytes([]byte{b})
}

func (ch *channel) WriteUint8(i uint8) error {
	return ch.WriteByte(byte(i))
}

func (ch *channel) WriteInt8(i int8) error {
	return ch.WriteUint8(uint8(i))
}

func (ch *channel) WriteUint16(i uint16) error {
	bs := make([]byte, 2)
	endian.PutUint16(bs, i)
	return ch.WriteBytes(bs)
}

func (ch *channel) WriteInt16(i int16) error {
	return ch.WriteUint16(uint16(i))
}

func (ch *channel) WriteUint32(i uint32) error {
	bs := make([]byte, 4)
	endian.PutUint32(bs, i)
	return ch.WriteBytes(bs)
}

func (ch *channel) WriteInt32(i int32) error {
	return ch.WriteUint32(uint32(i))
}

func (ch *channel) WriteUint64(i uint64) error {
	bs := make([]byte, 8)
	endian.PutUint64(bs, i)
	return ch.WriteBytes(bs)
}

func (ch *channel) WriteInt64(i int64) error {
	return ch.WriteUint64(uint64(i))
}

func (ch *channel) WriteString(s string) error {
	return ch.WriteBytes([]byte(s))
}
