package cgo

const (
	BIGDATETIME    ASEType = 35
	BIGINT         ASEType = 30
	BIGTIME        ASEType = 36
	BINARY         ASEType = 1
	BIT            ASEType = 11
	BLOB           ASEType = 26
	BOUNDARY       ASEType = 22
	CHAR           ASEType = 0
	DATE           ASEType = 27
	DATETIME       ASEType = 12
	DATETIME4      ASEType = 13
	DECIMAL        ASEType = 17
	FLOAT          ASEType = 10
	IMAGE          ASEType = 5
	IMAGELOCATOR   ASEType = 38
	INT            ASEType = 8
	LONG           ASEType = 20
	LONGBINARY     ASEType = 3
	LONGCHAR       ASEType = 2
	MONEY          ASEType = 14
	MONEY4         ASEType = 15
	NUMERIC        ASEType = 16
	REAL           ASEType = 9
	SENSITIVITY    ASEType = 21
	SMALLINT       ASEType = 7
	TEXT           ASEType = 4
	TEXTLOCATOR    ASEType = 37
	TIME           ASEType = 28
	TINYINT        ASEType = 6
	UBIGINT        ASEType = 33
	UINT           ASEType = 32
	UNICHAR        ASEType = 25
	UNITEXT        ASEType = 29
	UNITEXTLOCATOR ASEType = 39
	USER           ASEType = 100
	USHORT         ASEType = 24
	USMALLINT      ASEType = 31
	VARBINARY      ASEType = 19
	VARCHAR        ASEType = 18
	VOID           ASEType = 23
	XML            ASEType = 34
)

var type2string = map[ASEType]string{
	BIGDATETIME:    "BIGDATETIME",
	BIGINT:         "BIGINT",
	BIGTIME:        "BIGTIME",
	BINARY:         "BINARY",
	BIT:            "BIT",
	BLOB:           "BLOB",
	BOUNDARY:       "BOUNDARY",
	CHAR:           "CHAR",
	DATE:           "DATE",
	DATETIME:       "DATETIME",
	DATETIME4:      "DATETIME4",
	DECIMAL:        "DECIMAL",
	FLOAT:          "FLOAT",
	IMAGE:          "IMAGE",
	IMAGELOCATOR:   "IMAGELOCATOR",
	INT:            "INT",
	LONG:           "LONG",
	LONGBINARY:     "LONGBINARY",
	LONGCHAR:       "LONGCHAR",
	MONEY:          "MONEY",
	MONEY4:         "MONEY4",
	NUMERIC:        "NUMERIC",
	REAL:           "REAL",
	SENSITIVITY:    "SENSITIVITY",
	SMALLINT:       "SMALLINT",
	TEXT:           "TEXT",
	TEXTLOCATOR:    "TEXTLOCATOR",
	TIME:           "TIME",
	TINYINT:        "TINYINT",
	UBIGINT:        "UBIGINT",
	UINT:           "UINT",
	UNICHAR:        "UNICHAR",
	UNITEXT:        "UNITEXT",
	UNITEXTLOCATOR: "UNITEXTLOCATOR",
	USER:           "USER",
	USHORT:         "USHORT",
	USMALLINT:      "USMALLINT",
	VARBINARY:      "VARBINARY",
	VARCHAR:        "VARCHAR",
	VOID:           "VOID",
	XML:            "XML",
}
