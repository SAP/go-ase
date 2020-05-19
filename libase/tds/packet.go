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

func (packet Packet) Bytes() []byte {
	bs := make([]byte, int(packet.Header.Length))
	packet.Header.Read(bs[:MsgHeaderLength])
	copy(bs[MsgHeaderLength:], packet.Data)
	return bs
}

func (packet *Packet) ReadFrom(reader io.Reader) (int64, error) {
	var totalBytes int64

	packet.Header = MessageHeader{}
	n, err := packet.Header.ReadFrom(reader)
	if err != nil {
		// No check for io.EOF here - if there aren't enough bytes for
		// a header the parsing went wrong
		return n, fmt.Errorf("failed to read header: %v", err)
	}

	totalBytes += n

	packet.Data = make([]byte, packet.Header.Length-MsgHeaderLength)

	m, err := reader.Read(packet.Data)
	totalBytes += int64(m)

	if err != nil {
		if err == io.EOF {
			if packet.Header.MsgType == TDS_DONE {
				return totalBytes, io.EOF
			}
			return totalBytes, nil
		}
		return totalBytes, fmt.Errorf("error reading body: %v", err)
	}

	return totalBytes, nil
}

func (packet Packet) WriteTo(writer io.Writer) (int, error) {
	return writer.Write(packet.Bytes())
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
