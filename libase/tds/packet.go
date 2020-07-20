package tds

import (
	"fmt"
	"io"
)

// Packet represents a single packet in a message.
type Packet struct {
	Header MessageHeader
	Data   []byte
}

func NewPacket() *Packet {
	packet := &Packet{}
	packet.Header = NewMessageHeader()
	packet.Data = make([]byte, MsgBodyLength)
	return packet
}

func (packet Packet) Bytes() ([]byte, error) {
	bs := make([]byte, int(packet.Header.Length))

	_, err := packet.Header.Read(bs[:MsgHeaderLength])
	if err != nil {
		return nil, fmt.Errorf("error reading header into byte slice: %w", err)
	}

	copy(bs[MsgHeaderLength:], packet.Data)
	return bs, nil
}

func (packet *Packet) ReadFrom(reader io.Reader) (int64, error) {
	var totalBytes int64

	packet.Header = MessageHeader{}
	n, err := packet.Header.ReadFrom(reader)
	if err != nil {
		return n, fmt.Errorf("failed to read header: %w", err)
	}

	totalBytes += n

	packet.Data = make([]byte, packet.Header.Length-MsgHeaderLength)

	m, err := reader.Read(packet.Data)
	totalBytes += int64(m)

	if err != nil {
		if err == io.EOF {
			if packet.Header.MsgType == TDS_BUF_CLOSE {
				return totalBytes, io.EOF
			}
			return totalBytes, nil
		}
		return totalBytes, fmt.Errorf("error reading body: %w", err)
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
	return fmt.Sprintf(
		"Type: %d, Status: %d, Length: %d, Channel: %d, PacketNr: %d, Window: %d, DataLen: %d, Data: %#v",
		packet.Header.MsgType,
		packet.Header.Status,
		packet.Header.Length,
		packet.Header.Channel,
		packet.Header.PacketNr,
		packet.Header.Window,
		len(packet.Data),
		packet.Data,
	)
}
