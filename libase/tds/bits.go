package tds

// TODO better name
type TDSSecurity int

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

//go:generate stringer -type=TDSToken
type TDSToken byte

const (
	TDS_CURDECLARE3  TDSToken = 0x10
	TDS_PARAMFMT2             = 0x20
	TDS_LANGUAGE              = 0x21
	TDS_ORDERBY2              = 0x22
	TDS_CURDECLARE2           = 0x23
	TDS_COLFMTOLD             = 0x2A
	TDS_DEBUGCMD              = 0x60
	TDS_ROWFMT2               = 0x61
	TDS_DYNAMIC2              = 0x62
	TDS_MSG                   = 0x65
	TDS_LOGOUT                = 0x71
	TDS_OFFSET                = 0x78
	TDS_RETURNSTATUS          = 0x79
	TDS_PROCID                = 0x7C
	TDS_CURCLOSE              = 0x80
	TDS_CURDELETE             = 0x81
	TDS_CURFETCH              = 0x82
	TDS_CURINFO               = 0x83
	TDS_CUROPEN               = 0x84
	TDS_CURUPDATE             = 0x85
	TDS_CURDECLARE            = 0x86
	TDS_CURINFO2              = 0x87
	TDS_CURINFO3              = 0x88
	TDS_COLNAME               = 0xA0
	TDS_COLFMT                = 0xA1
	TDS_EVENTNOTICE           = 0xA2
	TDS_TABNAME               = 0xA4
	TDS_COLINFO               = 0xA5
	TDS_OPTIONCMD             = 0xA6
	TDS_ALTNAME               = 0xA7
	TDS_ALTFMT                = 0xA8
	TDS_ORDERBY               = 0xA9
	TDS_ERROR                 = 0xAA
	TDS_INFO                  = 0xAB
	TDS_RETURNVALUE           = 0xAC
	TDS_LOGINACK              = 0xAD
	TDS_CONTROL               = 0xAE
	TDS_ALTCONTROL            = 0xAF
	TDS_KEY                   = 0xCA
	TDS_ROW                   = 0xD1
	TDS_ALTROW                = 0xD3
	TDS_PARAMS                = 0xD7
	TDS_RPC                   = 0xE0
	TDS_CAPABILITY            = 0xE2
	TDS_ENVCHANGE             = 0xE3
	TDS_EED                   = 0xE5
	TDS_DBRPC                 = 0xE6
	TDS_DYNAMIC               = 0xE7
	TDS_DBRPC2                = 0xE8
	TDS_PARAMFMT              = 0xEC
	TDS_ROWFMT                = 0xEE
	TDS_DONE                  = 0xFD
	TDS_DONEPROC              = 0xFE
	TDS_DONEINPROC            = 0xFF
)
