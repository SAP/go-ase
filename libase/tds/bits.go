package tds

//go:generate stringer -type=TDSToken
type TDSToken byte

const (
	TDS_CURDECLARE3  TDSToken = 0x10
	TDS_PARAMFMT2    TDSToken = 0x20
	TDS_LANGUAGE     TDSToken = 0x21
	TDS_ORDERBY2     TDSToken = 0x22
	TDS_CURDECLARE2  TDSToken = 0x23
	TDS_COLFMTOLD    TDSToken = 0x2A
	TDS_DEBUGCMD     TDSToken = 0x60
	TDS_ROWFMT2      TDSToken = 0x61
	TDS_DYNAMIC2     TDSToken = 0x62
	TDS_MSG          TDSToken = 0x65
	TDS_LOGOUT       TDSToken = 0x71
	TDS_OFFSET       TDSToken = 0x78
	TDS_RETURNSTATUS TDSToken = 0x79
	TDS_PROCID       TDSToken = 0x7C
	TDS_CURCLOSE     TDSToken = 0x80
	TDS_CURDELETE    TDSToken = 0x81
	TDS_CURFETCH     TDSToken = 0x82
	TDS_CURINFO      TDSToken = 0x83
	TDS_CUROPEN      TDSToken = 0x84
	TDS_CURUPDATE    TDSToken = 0x85
	TDS_CURDECLARE   TDSToken = 0x86
	TDS_CURINFO2     TDSToken = 0x87
	TDS_CURINFO3     TDSToken = 0x88
	TDS_COLNAME      TDSToken = 0xA0
	TDS_COLFMT       TDSToken = 0xA1
	TDS_EVENTNOTICE  TDSToken = 0xA2
	TDS_TABNAME      TDSToken = 0xA4
	TDS_COLINFO      TDSToken = 0xA5
	TDS_OPTIONCMD    TDSToken = 0xA6
	TDS_ALTNAME      TDSToken = 0xA7
	TDS_ALTFMT       TDSToken = 0xA8
	TDS_ORDERBY      TDSToken = 0xA9
	TDS_ERROR        TDSToken = 0xAA
	TDS_INFO         TDSToken = 0xAB
	TDS_RETURNVALUE  TDSToken = 0xAC
	TDS_LOGINACK     TDSToken = 0xAD
	TDS_CONTROL      TDSToken = 0xAE
	TDS_ALTCONTROL   TDSToken = 0xAF
	TDS_KEY          TDSToken = 0xCA
	TDS_ROW          TDSToken = 0xD1
	TDS_ALTROW       TDSToken = 0xD3
	TDS_PARAMS       TDSToken = 0xD7
	TDS_RPC          TDSToken = 0xE0
	TDS_CAPABILITY   TDSToken = 0xE2
	TDS_ENVCHANGE    TDSToken = 0xE3
	TDS_EED          TDSToken = 0xE5
	TDS_DBRPC        TDSToken = 0xE6
	TDS_DYNAMIC      TDSToken = 0xE7
	TDS_DBRPC2       TDSToken = 0xE8
	TDS_PARAMFMT     TDSToken = 0xEC
	TDS_ROWFMT       TDSToken = 0xEE
	TDS_DONE         TDSToken = 0xFD
	TDS_DONEPROC     TDSToken = 0xFE
	TDS_DONEINPROC   TDSToken = 0xFF
)
