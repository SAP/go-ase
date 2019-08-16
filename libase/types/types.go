package types

import (
	"reflect"
	"time"
)

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
	ILLEGAL        ASEType = -1
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

var string2type = map[string]ASEType{
	"BIGDATETIME":    BIGDATETIME,
	"BIGINT":         BIGINT,
	"BIGTIME":        BIGTIME,
	"BINARY":         BINARY,
	"BIT":            BIT,
	"BLOB":           BLOB,
	"BOUNDARY":       BOUNDARY,
	"CHAR":           CHAR,
	"DATE":           DATE,
	"DATETIME":       DATETIME,
	"DATETIME4":      DATETIME4,
	"DECIMAL":        DECIMAL,
	"FLOAT":          FLOAT,
	"ILLEGAL":        ILLEGAL,
	"IMAGE":          IMAGE,
	"IMAGELOCATOR":   IMAGELOCATOR,
	"INT":            INT,
	"LONG":           LONG,
	"LONGBINARY":     LONGBINARY,
	"LONGCHAR":       LONGCHAR,
	"MONEY":          MONEY,
	"MONEY4":         MONEY4,
	"NUMERIC":        NUMERIC,
	"REAL":           REAL,
	"SENSITIVITY":    SENSITIVITY,
	"SMALLINT":       SMALLINT,
	"TEXT":           TEXT,
	"TEXTLOCATOR":    TEXTLOCATOR,
	"TIME":           TIME,
	"TINYINT":        TINYINT,
	"UBIGINT":        UBIGINT,
	"UINT":           UINT,
	"UNICHAR":        UNICHAR,
	"UNITEXT":        UNITEXT,
	"UNITEXTLOCATOR": UNITEXTLOCATOR,
	"USER":           USER,
	"USHORT":         USHORT,
	"USMALLINT":      USMALLINT,
	"VARBINARY":      VARBINARY,
	"VARCHAR":        VARCHAR,
	"VOID":           VOID,
	"XML":            XML,
}

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
	ILLEGAL:        "ILLEGAL",
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

var type2reflect = map[ASEType]reflect.Type{
	BIGDATETIME:    reflect.TypeOf(time.Time{}),
	BIGINT:         reflect.TypeOf(int64(0)),
	BIGTIME:        reflect.TypeOf(time.Time{}),
	BINARY:         reflect.SliceOf(reflect.TypeOf(byte(0))),
	BIT:            reflect.TypeOf(bool(false)),
	BLOB:           reflect.SliceOf(reflect.TypeOf(byte(0))),
	BOUNDARY:       nil,
	CHAR:           reflect.TypeOf(string("")),
	DATE:           reflect.TypeOf(time.Time{}),
	DATETIME:       reflect.TypeOf(time.Time{}),
	DATETIME4:      reflect.TypeOf(time.Time{}),
	DECIMAL:        reflect.TypeOf(&Decimal{}),
	FLOAT:          reflect.TypeOf(float64(0)),
	ILLEGAL:        nil,
	IMAGE:          reflect.SliceOf(reflect.TypeOf(byte(0))),
	IMAGELOCATOR:   nil,
	INT:            reflect.TypeOf(int32(0)),
	LONG:           nil,
	LONGBINARY:     reflect.SliceOf(reflect.TypeOf(byte(0))),
	LONGCHAR:       reflect.TypeOf(string("")),
	MONEY:          reflect.TypeOf(&Decimal{}),
	MONEY4:         reflect.TypeOf(&Decimal{}),
	NUMERIC:        reflect.TypeOf(&Decimal{}),
	REAL:           nil,
	SENSITIVITY:    nil,
	SMALLINT:       reflect.TypeOf(int16(0)),
	TEXT:           reflect.TypeOf(string("")),
	TEXTLOCATOR:    nil,
	TIME:           reflect.TypeOf(time.Time{}),
	TINYINT:        reflect.TypeOf(uint8(0)),
	UBIGINT:        reflect.TypeOf(uint64(0)),
	UINT:           reflect.TypeOf(uint32(0)),
	UNICHAR:        reflect.TypeOf(string("")),
	UNITEXT:        reflect.TypeOf(string("")),
	UNITEXTLOCATOR: nil,
	USER:           nil,
	USHORT:         reflect.TypeOf(uint16(0)),
	USMALLINT:      reflect.TypeOf(uint16(0)),
	VARBINARY:      reflect.SliceOf(reflect.TypeOf(byte(0))),
	VARCHAR:        reflect.TypeOf(string("")),
	VOID:           nil,
	XML:            reflect.SliceOf(reflect.TypeOf(byte(0))),
}

var type2interface = map[ASEType]interface{}{
	BIGDATETIME:    time.Time{},
	BIGINT:         int64(0),
	BIGTIME:        time.Time{},
	BINARY:         []byte{0},
	BIT:            bool(false),
	BLOB:           []byte{0},
	BOUNDARY:       nil,
	CHAR:           string(""),
	DATE:           time.Time{},
	DATETIME:       time.Time{},
	DATETIME4:      time.Time{},
	DECIMAL:        &Decimal{},
	FLOAT:          float64(0),
	ILLEGAL:        nil,
	IMAGE:          []byte{0},
	IMAGELOCATOR:   nil,
	INT:            int32(0),
	LONG:           nil,
	LONGBINARY:     []byte{0},
	LONGCHAR:       string(""),
	MONEY:          &Decimal{},
	MONEY4:         &Decimal{},
	NUMERIC:        &Decimal{},
	REAL:           nil,
	SENSITIVITY:    nil,
	SMALLINT:       int16(0),
	TEXT:           string(""),
	TEXTLOCATOR:    nil,
	TIME:           time.Time{},
	TINYINT:        uint8(0),
	UBIGINT:        uint64(0),
	UINT:           uint32(0),
	UNICHAR:        string(""),
	UNITEXT:        string(""),
	UNITEXTLOCATOR: nil,
	USER:           nil,
	USHORT:         uint16(0),
	USMALLINT:      uint16(0),
	VARBINARY:      []byte{0},
	VARCHAR:        string(""),
	VOID:           nil,
	XML:            []byte{0},
}
