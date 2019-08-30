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

	// // Create byte slice large enough for the remaining data
	// dataLen := int(packet.Header.Length) - MsgHeaderLength
	// log.Printf("dataLen: %d", dataLen)
	// if dataLen <= 0 {
	// 	if packet.Header.MsgType == TDS_DONE {
	// 		return totalBytes, io.EOF
	// 	}
	// 	return totalBytes, nil
	// }

	// Length in headers sent by the server isn't filled in, hence
	// MsgBodyLength is used
	packet.Data = make([]byte, packet.Header.Length)

	m, err := reader.Read(packet.Data)
	totalBytes += int64(m)

	// // TODO remove
	// err = ioutil.WriteFile(
	// 	fmt.Sprintf("/tmp/packet-%s.pack", time.Now().Format(time.RFC3339)),
	// 	packet.Data, 0755,
	// )
	// if err != nil {
	// 	return totalBytes, fmt.Errorf("failed to write packet data to file: %v", err)
	// }

	if err != nil {
		if err == io.EOF {
			if packet.Header.MsgType == TDS_DONE {
				return totalBytes, io.EOF
			}
			return totalBytes, nil
		}

		if err != nil {
			return totalBytes, fmt.Errorf("error reading body: %v", err)
		}

		// TODO remove, though a check would be nice - see MsgBodyLength
		// above
		// if m != MsgBodyLength {
		// 	return totalBytes, fmt.Errorf("not enough bytes read, expected %d, read %d",
		// 		len(packet.Data), m)
		// }

	}

	return totalBytes, nil
}

func (packet Packet) String() string {
	return fmt.Sprintf(
		"Type: %d, Status: %d, Length: %d, Channel: %d, PacketNr: %d, Window: %d, DataLen: %d",
		packet.Header.MsgType,
		packet.Header.Status,
		packet.Header.Length,
		packet.Header.Channel,
		packet.Header.PacketNr,
		packet.Header.Window,
		len(packet.Data),
	)
}
