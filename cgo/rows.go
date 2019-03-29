package cgo

//#include <stdlib.h>
//#include "ctlib.h"
//#include "bridge.h"
import "C"
import (
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/SAP/go-ase/libase/types"
)

// rows is the struct which represents a database result set
type rows struct {
	cmd *csCommand

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
	_ driver.Rows                           = (*rows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*rows)(nil)
	_ driver.RowsColumnTypeLength           = (*rows)(nil)
	_ driver.RowsColumnTypeNullable         = (*rows)(nil)
	_ driver.RowsColumnTypePrecisionScale   = (*rows)(nil)
	_ driver.RowsColumnTypeScanType         = (*rows)(nil)
)

func newRows(cmd *csCommand) (*rows, error) {
	var numCols C.CS_INT
	retval := C.ct_res_info(cmd.cmd, C.CS_NUMDATA, unsafe.Pointer(&numCols), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to read column count")
	}

	if numCols <= 0 {
		return nil, fmt.Errorf("Result set with zero columnes")
	}

	r := &rows{
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
		case C.CS_IMAGE_TYPE:
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

func (rows *rows) Close() error {
	for i := 0; i < rows.numCols; i++ {
		if rows.dataFmts[i] != nil {
			C.free(unsafe.Pointer(rows.dataFmts[i]))
		}
		if rows.colData[i] != nil {
			C.free(rows.colData[i])
		}
	}

	for r, _, err := rows.cmd.resultsHelper(); err != io.EOF; r, _, err = rows.cmd.resultsHelper() {
		if err != nil {
			return fmt.Errorf("Received error reading results: %v", err)
		}

		if r != nil {
			return fmt.Errorf("Received rows reading results, exiting: %v", r)
		}
	}

	if !rows.cmd.isDynamic {
		err := rows.cmd.drop()
		if err != nil {
			return fmt.Errorf("Error dropping command: %v", err)
		}
	}
	rows.cmd = nil

	return nil
}

func (rows *rows) Columns() []string {
	ret := make([]string, rows.numCols)

	for i, col := range rows.dataFmts {
		ret[i] = C.GoString(&((*C.CS_DATAFMT)(col).name[0]))
	}

	return ret
}

func (rows *rows) Next(dest []driver.Value) error {
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
		case types.BINARY:
			dest[i] = C.GoBytes(rows.colData[i], rows.dataFmts[i].maxlength)
		case types.INT:
			dest[i] = int64(int32(*((*C.CS_BIGINT)(rows.colData[i]))))
		case types.BIGINT:
			dest[i] = int64(*((*C.CS_BIGINT)(rows.colData[i])))
		case types.UINT:
			dest[i] = uint64(uint32(*((*C.CS_UBIGINT)(rows.colData[i]))))
		case types.UBIGINT:
			dest[i] = uint64(*((*C.CS_UBIGINT)(rows.colData[i])))
		case types.FLOAT:
			dest[i] = float64(*((*C.CS_FLOAT)(rows.colData[i])))
		case types.BIT:
			dest[i] = false
			if int(*(*C.CS_BIT)(rows.colData[i])) == 1 {
				dest[i] = true
			}
		case types.CHAR:
			dest[i] = C.GoString((*C.char)(rows.colData[i]))
		case types.BIGDATETIME:
			// TODO: convert from data into time.Time - e.g. cs_convert?
		default:
			return fmt.Errorf("Unhandled Go type: %+v", rows.colASEType[i])
		}
	}

	return nil
}

func (rows *rows) ColumnTypeDatabaseTypeName(index int) string {
	return rows.colASEType[index].String()
}

func (rows *rows) ColumnTypeNullable(index int) (bool, bool) {
	return (rows.dataFmts[index].status|C.CS_CANBENULL == C.CS_CANBENULL), true
}

func (rows *rows) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	if rows.dataFmts[index].scale == 0 && rows.dataFmts[index].precision == 9 {
		return 0, 0, false
	}

	return int64(rows.dataFmts[index].scale), int64(rows.dataFmts[index].precision), true
}

func (rows *rows) ColumnTypeLength(index int) (int64, bool) {
	// TODO
	switch rows.colASEType[index].GoType().(type) {
	default:
		return 0, false
	}
}

func (rows *rows) ColumnTypeScanType(index int) reflect.Type {
	return rows.colASEType[index].GoReflectType()
}
