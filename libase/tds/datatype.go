package tds

//go:generate stringer -type=DataType
type DataType uint8

const (
	// The TDS datatypes, ordered by tokenvalue.
	TDS_VOID         DataType = 0x1f
	TDS_IMAGE        DataType = 0x22
	TDS_TEXT         DataType = 0x23
	TDS_BLOB         DataType = 0x24
	TDS_VARBINARY    DataType = 0x25
	TDS_INTN         DataType = 0x26
	TDS_VARCHAR      DataType = 0x27
	TDS_BINARY       DataType = 0x2D
	TDS_CHAR         DataType = 0x2F
	TDS_INT1         DataType = 0x30
	TDS_DATE         DataType = 0x31
	TDS_BIT          DataType = 0x32
	TDS_TIME         DataType = 0x33
	TDS_INT2         DataType = 0x34
	TDS_INT4         DataType = 0x38
	TDS_SHORTDATE    DataType = 0x3A
	TDS_FLT4         DataType = 0x3B
	TDS_MONEY        DataType = 0x3C
	TDS_DATETIME     DataType = 0x3D
	TDS_FLT8         DataType = 0x3E
	TDS_UINT2        DataType = 0x41
	TDS_UINT4        DataType = 0x42
	TDS_UINT8        DataType = 0x43
	TDS_UINTN        DataType = 0x44
	TDS_SENSITIVITY  DataType = 0x67
	TDS_BOUNDARY     DataType = 0x68
	TDS_DECN         DataType = 0x6A
	TDS_NUMN         DataType = 0x6C
	TDS_FLTN         DataType = 0x6D
	TDS_MONEYN       DataType = 0x6E
	TDS_DATETIMN     DataType = 0x6F
	TDS_SHORTMONEY   DataType = 0x7A
	TDS_DATEN        DataType = 0x7B
	TDS_TIMEN        DataType = 0x93
	TDS_XML          DataType = 0xA3
	TDS_UNITEXT      DataType = 0xAE
	TDS_LONGCHAR     DataType = 0xAF
	TDS_BIGDATETIMEN DataType = 0xBB
	TDS_BIGTIMEN     DataType = 0xBC
	TDS_INT8         DataType = 0xBF
	TDS_LONGBINARY   DataType = 0xE1

	// Missing in tdspublic.h
	TDS_INTERVAL  DataType = 0x2e
	TDS_SINT1     DataType = 0xb0
	TDS_DATETIMEN DataType = 0x6f

	// TDS usertypes.
	TDS_USER_TEXT    DataType = 0x19
	TDS_USER_IMAGE   DataType = 0x20
	TDS_USER_UNITEXT DataType = 0x36
)
