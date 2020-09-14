// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package cgo

//#include <stdlib.h>
//#include "ctlib.h"
//#include "bridge.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"

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
	colASEType []ASEType
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
		colASEType: make([]ASEType, int(numCols)),
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
		asetype := (ASEType)(r.dataFmts[i].datatype)
		if asetype.String() == "" {
			r.Close()
			return nil, fmt.Errorf("Invalid ASEType: %v", r.dataFmts[i].datatype)
		}

		r.colASEType[i] = asetype

		// Set padding for datatypes that support it.
		switch r.dataFmts[i].datatype {
		case C.CS_BINARY_TYPE, C.CS_LONGBINARY_TYPE, C.CS_VARBINARY_TYPE, C.CS_CHAR_TYPE, C.CS_VARCHAR_TYPE, C.CS_LONGCHAR_TYPE:
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

	newRows, _, err := rows.cmd.ConsumeResponse(context.Background())
	if err != nil {
		return err
	}

	if newRows != nil {
		newRows.Close()
	}

	if !rows.cmd.isDynamic {
		err := rows.cmd.Drop()
		if err != nil {
			return fmt.Errorf("Error dropping command: %w", err)
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
		dataType := rows.colASEType[i].ToDataType()

		switch rows.colASEType[i] {
		case BIGINT, INT, SMALLINT, TINYINT, UBIGINT, UINT, USMALLINT, USHORT, FLOAT, REAL, BIT, MONEY, MONEY4, DATE, TIME, DATETIME, DATETIME4:
			bs := C.GoBytes(rows.colData[i], C.int(dataType.ByteSize()))
			val, err := dataType.GoValue(binary.LittleEndian, bs)
			if err != nil {
				return err
			}
			dest[i] = val
		case CHAR, VARCHAR, TEXT, LONGCHAR:
			dest[i] = C.GoString((*C.char)(rows.colData[i]))
		case BINARY, IMAGE:
			dest[i] = C.GoBytes(rows.colData[i], rows.dataFmts[i].maxlength)
		case DECIMAL, NUMERIC:
			csDec := (*C.CS_DECIMAL)(rows.colData[i])
			bs := C.GoBytes(
				unsafe.Pointer(&csDec.array),
				(C.int)(types.DecimalByteSize(int(csDec.precision))),
			)

			decI, err := dataType.GoValue(binary.LittleEndian, bs)
			if err != nil {
				return err
			}

			dec := decI.(*types.Decimal)
			dec.Precision = int(csDec.precision)
			dec.Scale = int(csDec.scale)
			dest[i] = dec
		case BIGDATETIME, BIGTIME:
			bs := C.GoBytes(rows.colData[i], 8)
			resp, err := dataType.GoValue(binary.LittleEndian, bs)
			if err != nil {
				return err
			}
			dest[i] = resp
		case UNICHAR, UNITEXT:
			b := C.GoBytes(rows.colData[i], rows.dataFmts[i].maxlength)
			s, err := dataType.GoValue(binary.LittleEndian, b)
			if err != nil {
				return err
			}
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
	case BINARY, IMAGE:
		return int64(rows.dataFmts[index].maxlength), true
	case CHAR:
		return int64(C.CS_MAX_CHAR), true
	default:
		return 0, false
	}
}

func (rows *Rows) ColumnTypeMaxLength(index int) int64 {
	return int64(rows.dataFmts[index].maxlength)
}

func (rows *Rows) ColumnTypeScanType(index int) reflect.Type {
	return rows.colASEType[index].ToDataType().GoReflectType()
}
