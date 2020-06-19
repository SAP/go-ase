package cgo

//#include <stdlib.h>
//#include "ctlib.h"
//#include "bridge.h"
import "C"
import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/SAP/go-ase/libase/asetime"
	"github.com/SAP/go-ase/libase/types"
)

// Rows is the struct which represents a database result set
type Rows struct {
	cmd *Command

	numCols int

	// dataFmts maps directly to columns in the result set. Each dataFmt
	// contains information regarding the column, such as the data type
	// or size.
	dataFmts []*C.CS_DATAFMT
	// colASEType maps directly to columns in the result set.
	// Each colASEType is the ASEType for the column.
	colASEType []types.ASEType
	// colData is a pointer to allocated memory according to the type
	// and size as indicated by the dataFmt.
	// the ctlibrary copies field data into this memory.
	colData []unsafe.Pointer
}

// Interface satisfaction checks
var (
	_ driver.Rows                           = (*Rows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*Rows)(nil)
	_ driver.RowsColumnTypeLength           = (*Rows)(nil)
	_ driver.RowsColumnTypeNullable         = (*Rows)(nil)
	_ driver.RowsColumnTypePrecisionScale   = (*Rows)(nil)
	_ driver.RowsColumnTypeScanType         = (*Rows)(nil)
)

func newRows(cmd *Command) (*Rows, error) {
	var numCols C.CS_INT
	retval := C.ct_res_info(cmd.cmd, C.CS_NUMDATA, unsafe.Pointer(&numCols), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to read column count")
	}

	if numCols <= 0 {
		return nil, fmt.Errorf("Result set with zero columnes")
	}

	r := &Rows{
		cmd:        cmd,
		numCols:    int(numCols),
		dataFmts:   make([]*C.CS_DATAFMT, int(numCols)),
		colASEType: make([]types.ASEType, int(numCols)),
		colData:    make([]unsafe.Pointer, int(numCols)),
	}

	// Setup column and row memory for ct to write into
	for i := 0; i < r.numCols; i++ {
		// Allocate columns dataFmt
		r.dataFmts[i] = (*C.CS_DATAFMT)(C.calloc(1, C.sizeof_CS_DATAFMT))

		// Fill dataFmt with information
		retval := C.ct_describe(cmd.cmd, (C.CS_INT)(i+1), r.dataFmts[i])
		if retval != C.CS_SUCCEED {
			r.Close()
			return nil, makeError(retval, "Failed to retrieve description of column")
		}

		// Set ASEType for column
		asetype := (types.ASEType)(r.dataFmts[i].datatype)
		if asetype.String() == "" {
			r.Close()
			return nil, fmt.Errorf("Invalid ASEType: %v", r.dataFmts[i].datatype)
		}

		r.colASEType[i] = asetype

		// Set padding for datatypes that support it.
		switch r.dataFmts[i].datatype {
		case C.CS_BINARY_TYPE, C.CS_LONGBINARY_TYPE, C.CS_VARBINARY_TYPE:
			r.dataFmts[i].format = C.CS_FMT_PADNULL
		case C.CS_XML_TYPE:
			r.dataFmts[i].format = C.CS_FMT_NULLTERM
		case C.CS_TEXT_TYPE, C.CS_UNITEXT_TYPE:
			r.dataFmts[i].format = C.CS_FMT_NULLTERM
		}

		// Allocate memory according maxlength of column
		r.colData[i] = C.calloc((C.ulong)(r.dataFmts[i].maxlength), C.sizeof_CS_BYTE)

		// Bind colData as the target for the data fetched with ct_fetch.
		// The last two arguments - copied and indicator - are optional.
		retval = C.ct_bind(cmd.cmd, (C.CS_INT)(i+1), r.dataFmts[i], r.colData[i], nil, nil)
		if retval != C.CS_SUCCEED {
			r.Close()
			return nil, makeError(retval, "Failed to bind data")
		}
	}

	return r, nil
}

func (rows *Rows) Close() error {
	for i := 0; i < rows.numCols; i++ {
		if rows.dataFmts[i] != nil {
			C.free(unsafe.Pointer(rows.dataFmts[i]))
		}
		if rows.colData[i] != nil {
			C.free(rows.colData[i])
		}
	}

	for r, _, _, err := rows.cmd.Response(); err != io.EOF; r, _, _, err = rows.cmd.Response() {
		if err != nil {
			return fmt.Errorf("Received error reading results: %v", err)
		}

		if r != nil {
			return fmt.Errorf("Received rows reading results, exiting: %v", r)
		}
	}

	if !rows.cmd.isDynamic {
		err := rows.cmd.Drop()
		if err != nil {
			return fmt.Errorf("Error dropping command: %v", err)
		}
	}
	rows.cmd = nil

	return nil
}

func (rows *Rows) Columns() []string {
	ret := make([]string, rows.numCols)

	for i, col := range rows.dataFmts {
		ret[i] = C.GoString(&((*C.CS_DATAFMT)(col).name[0]))
	}

	return ret
}

func (rows *Rows) Next(dest []driver.Value) error {
	retval := C.ct_fetch(rows.cmd.cmd, C.CS_UNUSED, C.CS_UNUSED, C.CS_UNUSED, nil)
	switch retval {
	case C.CS_SUCCEED:
		break
	case C.CS_END_DATA:
		return io.EOF
	case C.CS_ROW_FAIL, C.CS_FAIL:
		return makeError(retval, "Failed to retrieve rows")
	}

	if retval != C.CS_SUCCEED {
		return makeError(retval, "Failed to fetch next row")
	}

	for i := 0; i < len(rows.colData); i++ {
		switch rows.colASEType[i] {
		case types.BIGINT:
			dest[i] = int64(*((*C.CS_BIGINT)(rows.colData[i])))
		case types.INT:
			dest[i] = int32(*((*C.CS_INT)(rows.colData[i])))
		case types.SMALLINT:
			dest[i] = int16(*((*C.CS_SMALLINT)(rows.colData[i])))
		case types.TINYINT:
			dest[i] = uint8(*((*C.CS_TINYINT)(rows.colData[i])))
		case types.UBIGINT:
			dest[i] = uint64(*((*C.CS_UBIGINT)(rows.colData[i])))
		case types.UINT:
			dest[i] = uint32(*((*C.CS_UBIGINT)(rows.colData[i])))
		case types.USMALLINT, types.USHORT:
			dest[i] = uint16(*((*C.CS_USMALLINT)(rows.colData[i])))
		case types.DECIMAL, types.NUMERIC:
			csDec := (*C.CS_DECIMAL)(rows.colData[i])

			dec, err := types.NewDecimal(int(csDec.precision), int(csDec.scale))
			if err != nil {
				return fmt.Errorf("Received invalid precision/scale values from ASE: %v", err)
			}

			bs := C.GoBytes(unsafe.Pointer(&csDec.array), (C.int)(dec.ByteSize()))
			dec.SetBytes(bs[1:])

			if bs[0] == 0x1 {
				dec.Negate()
			}

			dest[i] = dec
		case types.FLOAT:
			dest[i] = float64(*((*C.CS_FLOAT)(rows.colData[i])))
		case types.REAL:
			dest[i] = float64(*((*C.CS_REAL)(rows.colData[i])))
		case types.MONEY:
			bs := C.GoBytes(rows.colData[i], 8)
			dec, err := types.NewDecimal(types.ASEMoneyPrecision, types.ASEMoneyScale)
			if err != nil {
				return fmt.Errorf("Received invalid precision/scale values from ASE: %v", err)
			}

			mnyhigh := binary.LittleEndian.Uint32(bs[:4])
			mnylow := binary.LittleEndian.Uint32(bs[4:])

			mny := int64(int64(mnyhigh)<<32 + int64(mnylow))

			dec.SetInt64(mny)

			dest[i] = dec
		case types.MONEY4:
			bs := C.GoBytes(rows.colData[i], 4)

			dec, err := types.NewDecimal(types.ASESmallMoneyPrecision, types.ASESmallMoneyScale)
			if err != nil {
				return fmt.Errorf("Received invalid precision/scale values from ASE: %v", err)
			}
			dec.SetInt64(int64(int32(binary.LittleEndian.Uint32(bs))))

			dest[i] = dec
		case types.DATE:
			b := C.GoBytes(rows.colData[i], 4)
			r := int32(binary.LittleEndian.Uint32(b))
			days := asetime.ASEDuration(r) * asetime.Day
			dest[i] = asetime.Epoch1900().AddDate(0, 0, days.Days())
		case types.TIME:
			b := C.GoBytes(rows.colData[i], 4)
			n := int(int32(binary.LittleEndian.Uint32(b)))
			dur := asetime.FractionalSecondToMillisecond(n)
			t := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
			dest[i] = t.Add(time.Duration(dur.Milliseconds()) * time.Millisecond)
		case types.DATETIME4:
			b := C.GoBytes(rows.colData[i], 4)

			days := binary.LittleEndian.Uint16(b[:2])
			mins := binary.LittleEndian.Uint16(b[2:])

			t := asetime.Epoch1900()
			t = t.AddDate(0, 0, int(days))
			t = t.Add(time.Duration(int(mins)) * time.Minute)

			dest[i] = t
		case types.DATETIME:
			b := C.GoBytes(rows.colData[i], 8)

			days := asetime.ASEDuration(int32(binary.LittleEndian.Uint32(b[:4]))) * asetime.Day
			s := asetime.FractionalSecondToMillisecond(int(binary.LittleEndian.Uint32(b[4:])))

			t := asetime.Epoch1900()
			t = t.AddDate(0, 0, days.Days())
			t = t.Add(time.Duration(s.Microseconds()) * time.Microsecond)

			dest[i] = t
		case types.BIGDATETIME:
			bs := C.GoBytes(rows.colData[i], 8)
			dur := asetime.ASEDuration(binary.LittleEndian.Uint64(bs))

			t := time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
			t = t.AddDate(0, 0, dur.Days())
			ms := dur.Microseconds() - (dur.Days() * int(asetime.Day))
			t = t.Add(time.Duration(ms) * time.Microsecond)

			dest[i] = t
		case types.BIGTIME:
			bs := C.GoBytes(rows.colData[i], 8)
			dur := asetime.ASEDuration(binary.LittleEndian.Uint64(bs))

			t := asetime.EpochRataDie()
			t = t.Add(time.Duration(dur) * time.Microsecond)

			dest[i] = t
		case types.CHAR, types.VARCHAR, types.TEXT, types.LONGCHAR:
			dest[i] = C.GoString((*C.char)(rows.colData[i]))
		case types.BINARY, types.IMAGE:
			dest[i] = C.GoBytes(rows.colData[i], rows.dataFmts[i].maxlength)
		case types.BIT:
			dest[i] = false
			if int(*(*C.CS_BIT)(rows.colData[i])) == 1 {
				dest[i] = true
			}

		case types.UNICHAR, types.UNITEXT:
			b := C.GoBytes(rows.colData[i], rows.dataFmts[i].maxlength)

			runes := []rune{}

			for i := 0; i < len(b); i++ {
				// Determine if byte is a utf16 surrogate - if so two
				// bytes must be consumed to form one utf16 code point
				if utf16.IsSurrogate(rune(b[i])) {
					r := utf16.DecodeRune(rune(b[i]), rune(b[i+1]))
					runes = append(runes, r)
					i++
				} else {
					runes = append(runes, rune(b[i]))
				}
			}

			s := string(runes)
			// Trim null bytes from the right - ASE always sends the
			// maximum bytes for the TEXT datatype, causing the string
			// to have a couple thousand null bytes. These are also
			// carried over in a string() conversion and cause
			// false-negatives in comparisons.
			s = strings.TrimRight(s, "\x00")

			dest[i] = s
		default:
			return fmt.Errorf("Unhandled Go type: %+v", rows.colASEType[i])
		}
	}

	return nil
}

func (rows *Rows) ColumnTypeDatabaseTypeName(index int) string {
	return rows.colASEType[index].String()
}

func (rows *Rows) ColumnTypeNullable(index int) (bool, bool) {
	return (rows.dataFmts[index].status|C.CS_CANBENULL == C.CS_CANBENULL), true
}

func (rows *Rows) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	if rows.dataFmts[index].scale == 0 && rows.dataFmts[index].precision == 9 {
		return 0, 0, false
	}

	return int64(rows.dataFmts[index].scale), int64(rows.dataFmts[index].precision), true
}

func (rows *Rows) ColumnTypeLength(index int) (int64, bool) {
	switch rows.colASEType[index] {
	case types.BINARY:
		return int64(rows.dataFmts[index].maxlength), true
	case types.CHAR:
		return int64(C.CS_MAX_CHAR), true
	default:
		return 0, false
	}
}

func (rows *Rows) ColumnTypeMaxLength(index int) int64 {
	return int64(rows.dataFmts[index].maxlength)
}

func (rows *Rows) ColumnTypeScanType(index int) reflect.Type {
	return rows.colASEType[index].GoReflectType()
}
