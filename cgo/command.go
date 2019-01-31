package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"context"
	"fmt"
	"io"
	"unsafe"
)

type csCommand struct {
	cmd *C.CS_COMMAND
}

// cancel cancels the current result set and drops the command.
//
// cancel automatically calls drop.
// cancel cannot be called after drop.
func (cmd *csCommand) cancel() error {
	retval := C.ct_cancel(nil, cmd.cmd, C.CS_CANCEL_ALL)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Error occurred while cancelling command")
	}

	return cmd.drop()
}

// drop finishes reading the results and drops the command.
//
// drop cannot be called after cancel.
func (cmd *csCommand) drop() error {
	retval := C.ct_cmd_drop(cmd.cmd)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Failed to drop command")
	}
	cmd.cmd = nil
	return nil
}

// exec allocates, prepares and sends a command.
//
// The return values are the command structure, a function to deallocate
// the command structure and an error, if any occurred.
func (conn *connection) exec(query string) (*csCommand, error) {
	cmd := &csCommand{}
	retval := C.ct_cmd_alloc(conn.conn, &cmd.cmd)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to allocate command structure")
	}

	sql := C.CString(query)
	defer C.free(unsafe.Pointer(sql))

	// Set language command
	retval = C.ct_command(cmd.cmd, C.CS_LANG_CMD, sql, C.CS_NULLTERM, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		cmd.drop()
		return nil, makeError(retval, "Failed to set language command")
	}

	// Send command to ASE
	retval = C.ct_send(cmd.cmd)
	if retval != C.CS_SUCCEED {
		cmd.drop()
		return nil, makeError(retval, "Failed to send command")
	}

	return cmd, nil
}

func (conn *connection) execContext(ctx context.Context, query string) (*csCommand, error) {
	recvCmd := make(chan *csCommand, 1)
	recvErr := make(chan error, 1)
	go func() {
		cmd, err := conn.exec(query)
		recvCmd <- cmd
		recvErr <- err
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case cmd := <-recvCmd:
			if cmd != nil {
				return cmd, nil
			}
		case err := <-recvErr:
			if err != nil {
				return nil, err
			}
		}
	}
}

// resultsHelper reads a single response from the command structure and
// handles it.
//
// When no more results are available this method returns io.EOF.
func (cmd *csCommand) resultsHelper() (*rows, *result, error) {
	var resultType C.CS_INT
	retval := C.ct_results(cmd.cmd, &resultType)

	switch retval {
	case C.CS_SUCCEED:
		// handle result type
		break
	case C.CS_END_RESULTS:
		return nil, nil, io.EOF // no more responses available, quit
	case C.CS_FAIL:
		return nil, nil, makeError(retval, "Command failed")
	default:
		return nil, nil, makeError(retval, "Invalid return code")
	}

	switch resultType {
	// fetchable results
	case C.CS_COMPUTE_RESULT, C.CS_CURSOR_RESULT, C.CS_PARAM_RESULT:
		fallthrough
	case C.CS_ROW_RESULT, C.CS_STATUS_RESULT:
		rows, err := newRows(cmd)
		if err != nil {
			return nil, nil, err
		}

		return rows, nil, nil

	// non-fetchable results
	case C.CS_COMPUTEFMT_RESULT:
		return nil, nil, nil
	case C.CS_MSG_RESULT:
		return nil, nil, nil
	case C.CS_ROWFMT_RESULT:
		return nil, nil, nil
	case C.CS_DESCRIBE_RESULT:
		return nil, nil, nil

	// other result types
	case C.CS_CMD_FAIL:
		err := cmd.cancel()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, makeError(retval, "Command failed, cancelled")
	case C.CS_CMD_DONE:
		var rowsAffected C.CS_INT
		retval := C.ct_res_info(cmd.cmd, C.CS_ROW_COUNT, unsafe.Pointer(&rowsAffected),
			C.CS_UNUSED, nil)
		if retval != C.CS_SUCCEED {
			return nil, nil, makeError(retval, "Failed to read affected rows")
		}

		return nil, &result{int64(rowsAffected)}, nil
	case C.CS_CMD_SUCCEED:
		return nil, nil, nil

	default:
		cmd.cancel()
		return nil, nil, fmt.Errorf("Unknown result type: %d", resultType)
	}
}

// results reads responses from the command structure until no more
// responses are available.
func (cmd *csCommand) results() (*rows, *result, error) {
	rows, err := cmd.resultsHelper()
	if err != nil {
		return nil, nil, err
	}

	if rows != nil {
		return rows, nil, nil
	}

	result := &result{}

	var rowsAffected C.long

	retval := C.ct_res_info(cmd.cmd, C.CS_ROW_COUNT, unsafe.Pointer(&rowsAffected), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		result.rowsAffected = -1
	} else {
		result.rowsAffected = int64(rowsAffected)
	}

	return nil, result, nil
}

func (cmd *csCommand) resultsContext(ctx context.Context) (*rows, *result, error) {
	recvRows := make(chan *rows, 1)
	recvResult := make(chan *result, 1)
	recvErr := make(chan error, 1)
	go func() {
		rows, result, err := cmd.results()
		recvErr <- err
		recvRows <- rows
		recvResult <- result
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case rows := <-recvRows:
			if rows != nil {
				return rows, nil, nil
			}
		case result := <-recvResult:
			if result != nil {
				return nil, result, nil
			}
		case err := <-recvErr:
			if err != nil {
				return nil, nil, err
			}
		}
	}
}
