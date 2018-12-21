package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"fmt"
	"unsafe"
)

// exec allocates, prepares and sends a command.
//
// The return values are the command structure, a function to deallocate
// the command structure and an error, if any occurred.
func (conn *connection) exec(query string) (*C.CS_COMMAND, func() error, error) {
	var cmd *C.CS_COMMAND
	retval := C.ct_cmd_alloc(conn.conn, &cmd)
	if retval != C.CS_SUCCEED {
		return nil, nil, makeError(retval, "Failed to allocate command structure")
	}

	cmdFree := func() error {

		// Read results once more - when executing language commands for
		// example the command structure won't be idle until ct_results
		// has been called three times. Once to confirm that the
		// language command executed successfully, the once to confirm
		// that there are no results and then once more - which returns
		// erroneous data but puts the command structure in the idle
		// state required to drop it.
		var resultType C.CS_INT
		C.ct_results(cmd, &resultType)

		retval = C.ct_cmd_drop(cmd)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "Failed to drop command")
		}

		return nil
	}

	sql := C.CString(query)
	defer C.free(unsafe.Pointer(sql))

	// Set language command
	retval = C.ct_command(cmd, C.CS_LANG_CMD, sql, C.CS_NULLTERM, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		cmdFree()
		return nil, nil, makeError(retval, "Failed to set language command")
	}

	// Send command to ASE
	retval = C.ct_send(cmd)
	if retval != C.CS_SUCCEED {
		cmdFree()
		return nil, nil, makeError(retval, "Failed to send command")
	}

	return cmd, cmdFree, nil
}

// resultsHelper reads a single response from the command structure and
// handles it.
//
// The function is designed to be used for the results function.
func resultsHelper(cmd *C.CS_COMMAND) (*rows, error) {
	var resultType C.CS_INT
	retval := C.ct_results(cmd, &resultType)

	switch retval {
	case C.CS_SUCCEED:
		// handle result type
		break
	case C.CS_END_RESULTS:
		return nil, nil // no more responses available, quit
	case C.CS_FAIL:
		return nil, makeError(retval, "Command failed")
	default:
		return nil, makeError(retval, "Invalid return code")
	}

	switch resultType {
	case C.CS_CMD_SUCCEED:
		// After CS_CMD_SUCCEED CS_CMD_DONE must be returned, hence the
		// resultsHelper is being called again.
		return resultsHelper(cmd)
	case C.CS_CMD_DONE:
		return nil, nil
	case C.CS_CMD_FAIL:
		retval = C.ct_cancel(nil, cmd, C.CS_CANCEL_ALL)
		return nil, makeError(retval, "Command failed, cancelled")
	case C.CS_ROW_RESULT:
		// TODO process these result types properly
		C.ct_cancel(nil, cmd, C.CS_CANCEL_ALL)
		return &rows{}, nil
	case C.CS_PARAM_RESULT, C.CS_STATUS_RESULT:
		// TODO process these result types properly
		C.ct_cancel(nil, cmd, C.CS_CANCEL_ALL)
		return &rows{}, nil
	default:
		C.ct_cancel(nil, cmd, C.CS_CANCEL_ALL)
		return nil, fmt.Errorf("Unknown result type: %d", resultType)
	}
}

// results reads responses from the command structure until no more
// responses are available.
func results(cmd *C.CS_COMMAND) (*rows, *result, error) {
	rows, err := resultsHelper(cmd)
	if err != nil {
		return nil, nil, err
	}

	if rows != nil {
		return rows, nil, nil
	}

	result := &result{}

	var rowsAffected C.long

	retval := C.ct_res_info(cmd, C.CS_ROW_COUNT, unsafe.Pointer(&rowsAffected), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		result.rowsAffected = -1
	} else {
		result.rowsAffected = int64(rowsAffected)
	}

	return nil, result, nil
}
