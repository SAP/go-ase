package tds

import "fmt"

//go:generate stringer -type=TDSMsgStatus
type TDSMsgStatus uint8

const (
	TDS_MSG_HASNOARGS TDSMsgStatus = iota
	TDS_MSG_HASARGS
)

// TODO better name
//go:generate stringer -type=TDSMsgId
type TDSMsgId uint16

type TDSOpaqueSecurityToken uint8

const (
	TDS_MSG_SEC_ENCRYPT TDSMsgId = iota + 1
	TDS_MSG_SEC_LOGPWD
	TDS_MSG_SEC_REMPWD
	TDS_MSG_SEC_CHALLENGE
	TDS_MSG_SEC_RESPONSE
	TDS_MSG_SEC_GETLABEL
	TDS_MSG_SEC_LABEL
	TDS_MSG_SQL_TBLNAME
	TDS_MSG_GW_RESERVED
	TDS_MSG_OMNI_CAPABILITIES
	TDS_MSG_SEC_OPAQUE
	TDS_MSG_HAFAILOVER
	TDS_MSG_EMPTY
	TDS_MSG_SEC_ENCRYPT2
	TDS_MSG_SEC_LOGPWD2
	TDS_MSG_SEC_SUP_CIPHER2
	TDS_MSG_MIG_REQ
	TDS_MSG_MIG_SYNC
	TDS_MSG_MIG_CONT
	TDS_MSG_MIG_IGN
	TDS_MSG_MIG_FAIL
	TDS_MSG_SEC_REMPWD2
	TDS_MSG_MIG_RESUME
	TDS_MSG_HELLO
	TDS_MSG_LOGINPARAMS
	TDS_MSG_GRID_MIGREQ
	TDS_MSG_GRID_QUIESCE
	TDS_MSG_GRID_UNQUIESCE
	TDS_MSG_GRID_EVENT
	TDS_MSG_SEC_ENCRYPT3
	TDS_MSG_SEC_LOGPWD3
	TDS_MSG_SEC_REMPWD3
	TDS_MSG_DR_MAP
	TDS_MSG_SEC_SYMKEY
	TDS_MSG_SEC_ENCRYPT4
)

const (
	/*
	 ** TDS_MSG_SEC_OPAQUE message types
	 */
	TDS_SEC_SECSESS TDSOpaqueSecurityToken = iota
	TDS_SEC_FORWARD
	TDS_SEC_SIGN
	TDS_SEC_OTHER
)

type MsgPackage struct {
	Status TDSMsgStatus
	MsgId  TDSMsgId
}

func NewMsgPackage(status TDSMsgStatus, msgId TDSMsgId) *MsgPackage {
	return &MsgPackage{
		Status: status,
		MsgId:  msgId,
	}
}

func (pkg *MsgPackage) ReadFrom(ch BytesChannel) error {
	var err error

	_, err = ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}

	var status uint8
	status, err = ch.Uint8()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.Status = (TDSMsgStatus)(status)

	msgId, err := ch.Uint16()
	if err != nil {
		return ErrNotEnoughBytes
	}
	pkg.MsgId = TDSMsgId(msgId)

	return nil
}

func (pkg MsgPackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteByte(byte(TDS_MSG))
	if err != nil {
		return err
	}

	// Length
	err = ch.WriteUint8(3)
	if err != nil {
		return err
	}

	err = ch.WriteUint8(uint8(pkg.Status))
	if err != nil {
		return err
	}

	return ch.WriteUint16(uint16(pkg.MsgId))
}

func (pkg MsgPackage) String() string {
	return fmt.Sprintf("%T(%s, %s)", pkg, pkg.Status, pkg.MsgId)
}
