package tds

import (
	"errors"
	"io"
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
			if err == io.EOF || err == ErrChannelExhausted {
				return i, io.EOF
			}
			return i, err
		}
	}

	return len(p), nil
}

// Read

func (ch *channel) Bytes(n int) ([]byte, error) {
	if n == 0 {
		return []byte{}, nil
	}

	bs := make([]byte, n)
	for n := range bs {
		// Keep reading until the channel is closed
		ok := true
		for ok {
			bs[n], ok = <-ch.ch
			if !ok {
				if ch.closed {
					return bs, ErrChannelClosed
				}
			} else {
				break
			}
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

// Write

func (ch *channel) WriteBytes(bs []byte) {
	for _, b := range bs {
		ch.ch <- b
	}
}

func (ch *channel) WriteByte(b byte) {
	ch.ch <- b
}

func (ch *channel) WriteUint8(i uint8) {
	ch.ch <- byte(i)
}

func (ch *channel) WriteUint16(i uint16) {
	bs := make([]byte, 2)
	endian.PutUint16(bs, i)
	ch.WriteBytes(bs)
}

func (ch *channel) WriteUint32(i uint32) {
	bs := make([]byte, 4)
	endian.PutUint32(bs, i)
	ch.WriteBytes(bs)
}

func (ch *channel) WriteUint64(i uint64) {
	bs := make([]byte, 8)
	endian.PutUint64(bs, i)
	ch.WriteBytes(bs)
}

func (ch *channel) WriteString(s string) {
	for _, r := range s {
		ch.ch <- byte(r)
	}
}
