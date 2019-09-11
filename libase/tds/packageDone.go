package tds

import "fmt"

//go:generate stringer -type=DoneState
type DoneState uint16

const (
	TDS_DONE_FINAL      DoneState = 0x0
	TDS_DONE_MORE                 = 0x1
	TDS_DONE_ERRROR               = 0x2
	TDS_DONE_INXACT               = 0x4
	TDS_DONE_PROC                 = 0x8
	TDS_DONE_COUNT                = 0x10
	TDS_DONE_ATTN                 = 0x20
	TDS_DONE_EVENT                = 0x40
	TDS_DONE_CUMULATIVE           = 0x80
)

//go:generate stringer -type=TransState
type TransState uint16

const (
	TDS_NOT_IN_TRAN TransState = iota
	TDS_TRAN_IN_PROGRESS
	TDS_TRAN_COMPLETED
	TDS_TRAN_FAIL
	TDS_TRAN_STMT_FAIL
)

type DonePackage struct {
	status    DoneState
	tranState TransState
	count     int32

	channelWrapper
}

func (pkg *DonePackage) ReadFrom(ch *channel) {
	pkg.ch = ch
	defer pkg.Finish()

	status, err := ch.Uint16()
	if err != nil {
		pkg.err = fmt.Errorf("failed to read done status: %w", err)
		return
	}
	pkg.status = DoneState(status)

	tranState, err := ch.Uint16()
	if err != nil {
		pkg.err = fmt.Errorf("failed to read done tran state: %w", err)
		return
	}
	pkg.tranState = TransState(tranState)

	if pkg.status|TDS_DONE_COUNT == TDS_DONE_COUNT {
		pkg.count, err = ch.Int32()
		if err != nil {
			pkg.err = fmt.Errorf("failed to read done count: %w", err)
			return
		}
	}
}

func (pkg DonePackage) WriteTo(ch *channel) error {
	err := ch.WriteUint16(uint16(pkg.status))
	if err != nil {
		return err
	}

	err = ch.WriteUint16(uint16(pkg.tranState))
	if err != nil {
		return err
	}

	if pkg.status|TDS_DONE_COUNT == TDS_DONE_COUNT {
		return ch.WriteInt32(pkg.count)
	}

	return nil
}

func (pkg DonePackage) String() string {
	return fmt.Sprintf("%s(%s)", pkg.status, pkg.tranState)
}
