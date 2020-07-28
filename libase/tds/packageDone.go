package tds

import (
	"fmt"
)

//go:generate stringer -type=DoneState
type DoneState uint16

const (
	TDS_DONE_FINAL      DoneState = 0x0
	TDS_DONE_MORE       DoneState = 0x1
	TDS_DONE_ERRROR     DoneState = 0x2
	TDS_DONE_INXACT     DoneState = 0x4
	TDS_DONE_PROC       DoneState = 0x8
	TDS_DONE_COUNT      DoneState = 0x10
	TDS_DONE_ATTN       DoneState = 0x20
	TDS_DONE_EVENT      DoneState = 0x40
	TDS_DONE_CUMULATIVE DoneState = 0x80
)

//go:generate stringer -type=TransState
type TransState uint16

const (
	TDS_NOT_IN_TRAN      TransState = 0x0
	TDS_TRAN_IN_PROGRESS TransState = 0x1
	TDS_TRAN_COMPLETED   TransState = 0x2
	TDS_TRAN_FAIL        TransState = 0x3
	TDS_TRAN_STMT_FAIL   TransState = 0x4
)

type DonePackage struct {
	Status    DoneState
	TranState TransState
	Count     int32
}

type DoneProcPackage struct{ DonePackage }
type DoneInProcPackage struct{ DonePackage }

func (pkg *DonePackage) ReadFrom(ch BytesChannel) error {
	status, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Status = DoneState(status)

	tranState, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.TranState = TransState(tranState)

	pkg.Count, err = ch.Int32()
	if err != nil {
		return ErrNotEnoughBytes
	}

	return nil
}

func (pkg DonePackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteByte(byte(TDS_DONE))
	if err != nil {
		return err
	}

	err = ch.WriteUint16(uint16(pkg.Status))
	if err != nil {
		return err
	}

	err = ch.WriteUint16(uint16(pkg.TranState))
	if err != nil {
		return err
	}

	return ch.WriteInt32(pkg.Count)
}

func (pkg DonePackage) String() string {
	stati := deBitmask(int(pkg.Status), int(TDS_DONE_CUMULATIVE))
	strStati := ""
	if len(stati) == 0 {
		strStati = TDS_DONE_FINAL.String()
	} else {
		for i, status := range stati {
			strStati += DoneState(status).String()
			if i+1 != len(stati) {
				strStati += "|"
			}
		}
	}

	transi := deBitmask(int(pkg.TranState), int(TDS_TRAN_STMT_FAIL))
	strTransi := ""
	if len(transi) == 0 {
		strTransi = TDS_NOT_IN_TRAN.String()
	} else {
		for i, trans := range transi {
			strTransi += TransState(trans).String()
			if i+1 != len(transi) {
				strTransi += "|"
			}
		}
	}

	return fmt.Sprintf("%T(%s, %s)", pkg, strStati, strTransi)
}
