package tds

import "fmt"

//go:generate stringer -type=TDSMsgStatus
type TDSMsgStatus uint8

const (
	TDS_MSG_HASNOARGS TDSMsgStatus = iota
	TDS_MSG_HASARGS
)

// TODO better name
//go:generate stringer -type=TDSSecurity
type TDSSecurity uint8

const (
	TDS_MSG_SEC_ENCRYPT TDSSecurity = iota
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

	TDS_MSG_SEC_ENCRYPT3 = iota + 30
	TDS_MSG_SEC_LOGPWD3
	TDS_MSG_SEC_REMPWD3
	TDS_MSG_DR_MAP
	TDS_MSG_SEC_SYMKEY
	TDS_MSG_SEC_ENCRYPT4

	/*
	 ** TDS_MSG_SEC_OPAQUE message types
	 */
	TDS_SEC_SECSESS = iota
	TDS_SEC_FORWARD
	TDS_SEC_SIGN
	TDS_SEC_OTHER
)

type MsgPackage struct {
	Length uint8
	Status TDSMsgStatus
	MsgId  uint16
}

func (pkg *MsgPackage) ReadFrom(ch *channel) error {
	var err error

	pkg.Length, err = ch.Uint8()
	if err != nil {
		return err
	}

	var status uint8
	status, err = ch.Uint8()
	if err != nil {
		return err
	}
	pkg.Status = (TDSMsgStatus)(status)

	pkg.MsgId, err = ch.Uint16()
	return err
}

func (pkg MsgPackage) WriteTo(ch *channel) error {
	err := ch.WriteUint8(pkg.Length)
	if err != nil {
		return err
	}

	err = ch.WriteUint8(uint8(pkg.Status))
	if err != nil {
		return err
	}

	return ch.WriteUint16(pkg.MsgId)
}

func (pkg MsgPackage) String() string {
	return fmt.Sprintf("%s(%d)", pkg.Status, pkg.MsgId)
}
