package tds

import (
	"errors"
)

var (
	ErrChannelClosed    = errors.New("channel is closed")
	ErrChannelExhausted = errors.New("channel is exhausted")
)

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

func (cr *channel) Close() {
	cr.closed = true
	close(cr.ch)
}

// Read satisfies the io.Reader interface
func (cr *channel) Read(p []byte) (int, error) {
	bs, err := cr.Bytes(len(p))

	for n, b := range bs {
		p[n] = b
	}

	return len(bs), err
}

// Read

func (cr *channel) Bytes(n int) ([]byte, error) {
	if n == 0 {
		return []byte{}, nil
	}

	bs := make([]byte, n)
	for n := range bs {
		// Keep reading until the channel is closed
		ok := true
		for ok {
			bs[n], ok = <-cr.ch
			if !ok {
				if cr.closed {
					return bs, ErrChannelClosed
				}
			} else {
				break
			}
		}
	}

	return bs, nil
}

func (cr *channel) Byte() (byte, error) {
	bs, err := cr.Bytes(1)
	return bs[0], err
}

func (cr *channel) Uint8() (uint8, error) {
	b, err := cr.Byte()
	return uint8(b), err
}

func (cr *channel) Int8() (int8, error) {
	b, err := cr.Byte()
	return int8(b), err
}

func (cr *channel) Uint16() (uint16, error) {
	bs, err := cr.Bytes(2)
	return endian.Uint16(bs), err
}

func (cr *channel) Int16() (int16, error) {
	i, err := cr.Uint16()
	return int16(i), err
}

func (cr *channel) Uint32() (uint32, error) {
	bs, err := cr.Bytes(4)
	return endian.Uint32(bs), err
}

func (cr *channel) Int32() (int32, error) {
	i, err := cr.Uint32()
	return int32(i), err
}

func (cr *channel) Uint64() (uint64, error) {
	bs, err := cr.Bytes(8)
	return endian.Uint64(bs), err
}

func (cr *channel) Int64() (int64, error) {
	i, err := cr.Uint64()
	return int64(i), err
}

func (cr *channel) String(size int) (string, error) {
	bs, err := cr.Bytes(size)
	return string(bs), err
}

// Write

func (cr *channel) WriteBytes(bs []byte) {
	for _, b := range bs {
		cr.ch <- b
	}
}

func (cr *channel) WriteByte(b byte) {
	cr.ch <- b
}

func (cr *channel) WriteUint8(i uint8) {
	cr.ch <- byte(i)
}

func (cr *channel) WriteUint16(i uint16) {
	bs := make([]byte, 2)
	endian.PutUint16(bs, i)
	cr.WriteBytes(bs)
}

func (cr *channel) WriteUint32(i uint32) {
	bs := make([]byte, 4)
	endian.PutUint32(bs, i)
	cr.WriteBytes(bs)
}

func (cr *channel) WriteUint64(i uint64) {
	bs := make([]byte, 8)
	endian.PutUint64(bs, i)
	cr.WriteBytes(bs)
}

func (cr *channel) WriteString(s string) {
	for _, r := range s {
		cr.ch <- byte(r)
	}
}
