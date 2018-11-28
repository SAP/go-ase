package driver

//#include "ctpublic.h"
import "C"
import (
	"fmt"
	"strconv"
)

// makeError creates an error out of the return code of an ASE routine
// and a message.
//
// The passed message leads the error and can thus be used to add
// specific information before the generic ASE error message.
//
// The return code is turned into a message according to the Client
// Library Programmers Guide.
func makeError(retcode C.CS_RETCODE, message string) error {
	// TODO parse retcode as integer and provide a map mapping error
	// codes to strings to consumers in addition to MakeError.
	var s string
	switch retcode {
	case C.CS_FAIL:
		s = "Routine failed"
	case C.CS_CANCELED:
		s = "Routine was canceled"
	case C.CS_PENDING:
		s = "Operation is pending, see asynchronous programming"
	case C.CS_BUSY:
		s = "An operation is already pending for this connection, see asynchronous programming"
	default:
		s = "Unknown error code: " + strconv.FormatInt(int64(retcode), 10)
	}

	return fmt.Errorf("%s: %s", message, s)
}
