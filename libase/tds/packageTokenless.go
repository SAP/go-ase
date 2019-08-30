package tds

import "bytes"

type TokenlessPackage struct {
	Data     *bytes.Buffer
	err      error
	finished bool
}

func (pkg *TokenlessPackage) ReadFrom(ch *channel) {
	_, pkg.err = pkg.Data.ReadFrom(ch)
	pkg.finished = true
}

func (pkg TokenlessPackage) Error() error {
	return pkg.err
}

func (pkg TokenlessPackage) Finished() bool {
	return pkg.finished
}

func (pack TokenlessPackage) Packets() chan Packet {
	// TODO configurable?
	ch := make(chan Packet, 10)

	header := MessageHeader{
		MsgType: TDS_BUF_LOGIN,
		Length:  MsgLength,
	}

	go func() {
		for pack.Data.Len() > 0 {
			// If the remaining buffer is smaller than the allowed
			// message length set length according to buffer length
			if pack.Data.Len() < MsgBodyLength {
				header.Length = uint16(pack.Data.Len() + MsgHeaderLength)
				// Last packet, send EOM flag
				header.Status = TDS_BUFSTAT_EOM
			}

			// // Increase PacketNr
			// header.PacketNr++

			p := Packet{
				Header: header,
				Data:   make([]byte, header.Length-MsgHeaderLength),
			}

			pack.Data.Read(p.Data)

			ch <- p
		}

		close(ch)
	}()

	return ch
}

func (pkg *TokenlessPackage) Write(bs []byte) (int, error) {
	return pkg.Data.Write(bs)
}

// TODO
func (pkg TokenlessPackage) String() string {
	return ""
}
