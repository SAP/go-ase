package tds

//go:generate stringer -type=Token
type Token byte

const (
	TDS_CURDECLARE3  Token = 0x10
	TDS_PARAMFMT2    Token = 0x20
	TDS_LANGUAGE     Token = 0x21
	TDS_ORDERBY2     Token = 0x22
	TDS_CURDECLARE2  Token = 0x23
	TDS_COLFMTOLD    Token = 0x2A
	TDS_DEBUGCMD     Token = 0x60
	TDS_ROWFMT2      Token = 0x61
	TDS_DYNAMIC2     Token = 0x62
	TDS_MSG          Token = 0x65
	TDS_LOGOUT       Token = 0x71
	TDS_OFFSET       Token = 0x78
	TDS_RETURNSTATUS Token = 0x79
	TDS_PROCID       Token = 0x7C
	TDS_CURCLOSE     Token = 0x80
	TDS_CURDELETE    Token = 0x81
	TDS_CURFETCH     Token = 0x82
	TDS_CURINFO      Token = 0x83
	TDS_CUROPEN      Token = 0x84
	TDS_CURUPDATE    Token = 0x85
	TDS_CURDECLARE   Token = 0x86
	TDS_CURINFO2     Token = 0x87
	TDS_CURINFO3     Token = 0x88
	TDS_COLNAME      Token = 0xA0
	TDS_COLFMT       Token = 0xA1
	TDS_EVENTNOTICE  Token = 0xA2
	TDS_TABNAME      Token = 0xA4
	TDS_COLINFO      Token = 0xA5
	TDS_OPTIONCMD    Token = 0xA6
	TDS_ALTNAME      Token = 0xA7
	TDS_ALTFMT       Token = 0xA8
	TDS_ORDERBY      Token = 0xA9
	TDS_ERROR        Token = 0xAA
	TDS_INFO         Token = 0xAB
	TDS_RETURNVALUE  Token = 0xAC
	TDS_LOGINACK     Token = 0xAD
	TDS_CONTROL      Token = 0xAE
	TDS_ALTCONTROL   Token = 0xAF
	TDS_KEY          Token = 0xCA
	TDS_ROW          Token = 0xD1
	TDS_ALTROW       Token = 0xD3
	TDS_PARAMS       Token = 0xD7
	TDS_RPC          Token = 0xE0
	TDS_CAPABILITY   Token = 0xE2
	TDS_ENVCHANGE    Token = 0xE3
	TDS_EED          Token = 0xE5
	TDS_DBRPC        Token = 0xE6
	TDS_DYNAMIC      Token = 0xE7
	TDS_DBRPC2       Token = 0xE8
	TDS_PARAMFMT     Token = 0xEC
	TDS_ROWFMT       Token = 0xEE
	TDS_DONE         Token = 0xFD
	TDS_DONEPROC     Token = 0xFE
	TDS_DONEINPROC   Token = 0xFF
)
