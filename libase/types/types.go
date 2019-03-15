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
	BIT:            reflect.TypeOf(uint64(0)),
	BLOB:           reflect.SliceOf(reflect.TypeOf(byte(0))),
	BOUNDARY:       nil,
	CHAR:           reflect.TypeOf(string("")),
	DATE:           reflect.SliceOf(reflect.TypeOf(byte(0))),
	DATETIME:       reflect.TypeOf(time.Time{}),
	DATETIME4:      reflect.SliceOf(reflect.TypeOf(byte(0))),
	DECIMAL:        reflect.TypeOf(float64(0)),
	FLOAT:          reflect.TypeOf(float64(0)),
	ILLEGAL:        nil,
	IMAGE:          reflect.SliceOf(reflect.TypeOf(byte(0))),
	IMAGELOCATOR:   nil,
	INT:            reflect.TypeOf(int64(0)),
	LONG:           reflect.TypeOf(uint64(0)),
	LONGBINARY:     reflect.SliceOf(reflect.TypeOf(byte(0))),
	LONGCHAR:       reflect.TypeOf(string("")),
	MONEY:          reflect.TypeOf(uint64(0)),
	MONEY4:         reflect.TypeOf(uint64(0)),
	NUMERIC:        reflect.TypeOf(uint64(0)),
	REAL:           nil,
	SENSITIVITY:    nil,
	SMALLINT:       reflect.TypeOf(int64(0)),
	TEXT:           reflect.TypeOf(string("")),
	TEXTLOCATOR:    nil,
	TIME:           reflect.SliceOf(reflect.TypeOf(byte(0))),
	TINYINT:        reflect.TypeOf(int64(0)),
	UBIGINT:        reflect.TypeOf(uint64(0)),
	UINT:           reflect.TypeOf(uint64(0)),
	UNICHAR:        reflect.TypeOf(string("")),
	UNITEXT:        reflect.TypeOf(string("")),
	UNITEXTLOCATOR: nil,
	USER:           nil,
	USHORT:         reflect.TypeOf(uint64(0)),
	USMALLINT:      reflect.TypeOf(uint64(0)),
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
	BIT:            uint64(0),
	BLOB:           []byte{0},
	BOUNDARY:       nil,
	CHAR:           string(""),
	DATE:           []byte{0},
	DATETIME:       time.Time{},
	DATETIME4:      []byte{0},
	DECIMAL:        float64(0),
	FLOAT:          float64(0),
	ILLEGAL:        nil,
	IMAGE:          []byte{0},
	IMAGELOCATOR:   nil,
	INT:            int64(0),
	LONG:           uint64(0),
	LONGBINARY:     []byte{0},
	LONGCHAR:       string(""),
	MONEY:          uint64(0),
	MONEY4:         uint64(0),
	NUMERIC:        uint64(0),
	REAL:           nil,
	SENSITIVITY:    nil,
	SMALLINT:       int64(0),
	TEXT:           string(""),
	TEXTLOCATOR:    nil,
	TIME:           []byte{0},
	TINYINT:        int64(0),
	UBIGINT:        uint64(0),
	UINT:           uint64(0),
	UNICHAR:        string(""),
	UNITEXT:        string(""),
	UNITEXTLOCATOR: nil,
	USER:           nil,
	USHORT:         uint64(0),
	USMALLINT:      uint64(0),
	VARBINARY:      []byte{0},
	VARCHAR:        string(""),
	VOID:           nil,
	XML:            []byte{0},
}
