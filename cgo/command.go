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

type Command struct {
	cmd       *C.CS_COMMAND
	isDynamic bool
}

// cancel cancels the current result set.
func (cmd *Command) Cancel() error {
	retval := C.ct_cancel(nil, cmd.cmd, C.CS_CANCEL_ALL)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Error occurred while cancelling command")
	}

	return nil
}

// drop deallocates the command.
func (cmd *Command) Drop() error {
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
func (conn *Connection) exec(query string) (*Command, error) {
	cmd := &Command{}
	retval := C.ct_cmd_alloc(conn.conn, &cmd.cmd)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to allocate command structure")
	}

	sql := C.CString(query)
	defer C.free(unsafe.Pointer(sql))

	// Set language command
	retval = C.ct_command(cmd.cmd, C.CS_LANG_CMD, sql, C.CS_NULLTERM, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		cmd.Drop()
		return nil, makeError(retval, "Failed to set language command")
	}

	// Send command to ASE
	retval = C.ct_send(cmd.cmd)
	if retval != C.CS_SUCCEED {
		cmd.Drop()
		return nil, makeError(retval, "Failed to send command")
	}

	return cmd, nil
}

func (conn *Connection) GenericExec(ctx context.Context, query string) (*Command, error) {
	recvCmd := make(chan *Command, 1)
	recvErr := make(chan error, 1)
	go func() {
		cmd, err := conn.exec(query)
		recvCmd <- cmd
		close(recvCmd)
		recvErr <- err
		close(recvErr)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-recvErr:
			cmd := <-recvCmd
			return cmd, err
		}
	}
}

// dynamic initializes a Command as a prepared statement.
func (conn *Connection) Dynamic(name string, query string) (*Command, error) {
	cmd := &Command{}
	cmd.isDynamic = true
	retval := C.ct_cmd_alloc(conn.conn, &cmd.cmd)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to allocate command structure")
	}

	// Initialize dynamic command
	q := C.CString(query)
	defer C.free(unsafe.Pointer(q))

	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	retval = C.ct_dynamic(cmd.cmd, C.CS_PREPARE, n, C.CS_NULLTERM, q, C.CS_NULLTERM)
	if retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to initialize dynamic command")
	}

	// Send command to ASE
	retval = C.ct_send(cmd.cmd)
	if retval != C.CS_SUCCEED {
		cmd.Drop()
		return nil, makeError(retval, "Failed to send command")
	}

	return cmd, nil
}

// Response reads a single response from the command structure and
// handles it.
//
// When no more results are available this method returns io.EOF.
//
// The third return value is a CS_INT, which may be the CS_RETCODE of
// ct_results when the command failed or finished or the result type if
// the result set requires further processing.
func (cmd *Command) Response() (*Rows, *Result, C.CS_INT, error) {
	var resultType C.CS_INT
	retval := C.ct_results(cmd.cmd, &resultType)

	switch retval {
	case C.CS_SUCCEED:
		// handle result type
		break
	case C.CS_END_RESULTS:
		return nil, nil, retval, io.EOF // no more responses available, quit
	case C.CS_FAIL:
		cmd.Cancel()
		return nil, nil, retval, makeError(retval, "Command failed")
	default:
		cmd.Cancel()
		return nil, nil, retval, makeError(retval, "Invalid return code")
	}

	switch resultType {
	// fetchable results
	case C.CS_COMPUTE_RESULT, C.CS_CURSOR_RESULT, C.CS_PARAM_RESULT:
		fallthrough
	case C.CS_ROW_RESULT, C.CS_STATUS_RESULT:
		rows, err := newRows(cmd)
		if err != nil {
			cmd.Cancel()
			return nil, nil, C.CS_UNUSED, err
		}

		return rows, nil, C.CS_UNUSED, nil

	// non-fetchable results
	case C.CS_COMPUTEFMT_RESULT, C.CS_MSG_RESULT, C.CS_ROWFMT_RESULT, C.CS_DESCRIBE_RESULT:
		return nil, nil, resultType, nil

	// other result types
	case C.CS_CMD_FAIL:
		cmd.Cancel()
		return nil, nil, C.CS_UNUSED, makeError(retval, "Command failed, cancelled")
	case C.CS_CMD_DONE:
		var rowsAffected C.CS_INT
		retval := C.ct_res_info(cmd.cmd, C.CS_ROW_COUNT, unsafe.Pointer(&rowsAffected),
			C.CS_UNUSED, nil)
		if retval != C.CS_SUCCEED {
			cmd.Cancel()
			return nil, nil, C.CS_UNUSED, makeError(retval, "Failed to read affected rows")
		}

		return nil, &Result{int64(rowsAffected)}, C.CS_UNUSED, nil
	case C.CS_CMD_SUCCEED:
		return nil, nil, C.CS_UNUSED, nil

	default:
		cmd.Cancel()
		return nil, nil, resultType, fmt.Errorf("Unknown result type: %d", resultType)
	}
}
