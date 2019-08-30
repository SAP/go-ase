package tds

import "fmt"

type EEDStatus uint8

const (
	TDS_NO_EED      EEDStatus = 0x00
	TDS_EED_FOLLOWS           = 0x1
	TDS_EED_INFO              = 0x2
)

type EEDPackage struct {
	Length       uint16
	MsgNumber    uint32
	State        uint8
	Class        uint8
	SQLStateLen  uint8
	SQLState     []byte
	Status       uint8
	TranState    uint16
	MsgLength    uint16
	Msg          []byte
	ServerLength uint8
	ServerName   []byte
	ProcLength   uint8
	ProcName     []byte
	LineNr       uint16
}

func (pkg *EEDPackage) Write(bs []byte) (int, error) {
	pkg.Length = endian.Uint16(bs[:2])
	pkg.MsgNumber = endian.Uint32(bs[2:6])

	pkg.SQLStateLen = bs[6]
	offset := 7

	pkg.SQLState = make([]byte, pkg.SQLStateLen)
	copy(pkg.SQLState, bs[offset:offset])

	offset += int(uint(pkg.SQLStateLen))

	pkg.Status = bs[offset]
	offset++

	pkg.TranState = endian.Uint16(bs[offset : offset+2])
	offset += 2

	pkg.MsgLength = endian.Uint16(bs[offset : offset+2])
	offset += 2

	pkg.Msg = make([]byte, pkg.MsgLength)
	copy(pkg.Msg, bs[offset:])
	offset += int(uint(pkg.MsgLength))

	pkg.ServerLength = bs[offset]
	offset++

	pkg.ServerName = make([]byte, pkg.ServerLength)
	copy(pkg.ServerName, bs[offset:])
	offset += int(uint(pkg.ServerLength))

	pkg.ProcLength = bs[offset]
	offset++

	pkg.ProcName = make([]byte, pkg.ProcLength)
	copy(pkg.ProcName, bs[offset:])
	offset += int(uint(pkg.ProcLength))

	pkg.LineNr = endian.Uint16(bs[offset : offset+2])
	offset += 2

	if offset != len(bs) {
		return offset, fmt.Errorf("failed to consume all input, %d left",
			len(bs)-offset,
		)
	}

	return offset, nil
}

func (pkg EEDPackage) String() string {
	return ""
}

func (pkg EEDPackage) Packets() chan Packet {
	return nil
}
