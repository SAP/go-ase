package cgo

//#include "ctlib.h"
//#include "bridge.h"
import "C"
import (
	"sync"

	libdsn "github.com/SAP/go-ase/libase/dsn"
)

// context wraps C.CS_CONTEXT to ensure that the context is being closed
// and deallocated after the last connection was closed.
type csContext struct {
	ctx  *C.CS_CONTEXT
	lock sync.Mutex

	// connections is a counter that keeps track of the number of
	// connections using the context to communicate with an ASE
	// instance.
	connections uint
}

// init performs two actions based on the state of context.ctx:
// context.ctx is not nil: Increment connection counter and return
// context.ctx is nil: Allocated and initialize context.ctx and set
//   connection counter to 1.
//
// Connections using a context must call this method before proceeding.
func (context *csContext) init() error {
	context.lock.Lock()
	defer context.lock.Unlock()

	// Return early if the context has been initialized
	if context.ctx != nil {
		// Increase connection counter
		context.connections += 1
		return nil
	}

	retval := C.cs_ctx_alloc(C.CS_CURRENT_VERSION, &context.ctx)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.cs_ctx_alloc failed")
	}

	retval = C.ct_init(context.ctx, C.CS_CURRENT_VERSION)
	if retval != C.CS_SUCCEED {
		err := context.drop()
		if err != nil {
			return err
		}
		return makeError(retval, "C.ct_init failed")
	}

	// Initialized context, set connection count.
	context.connections = 1
	return nil
}

// drop decrements the connections counter and deallocates the context
// if the counter dropped to zero.
func (context *csContext) drop() error {
	context.lock.Lock()
	defer context.lock.Unlock()

	// Decrease connection count and return early if there are still
	// connections using the context.
	context.connections -= 1
	if context.connections > 0 {
		return nil
	}

	// No connections using the context left, proceed with deallocation.

	retval := C.ct_exit(context.ctx, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_exit failed, has results pending")
	}

	retval = C.cs_ctx_drop(context.ctx)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.cs_ctx_drop frailed")
	}

	context.ctx = nil
	return nil
}

func (context *csContext) applyDSN(dsn libdsn.DsnInfo) error {
	if dsn.Prop("cgo-callback-client") == "yes" {
		retval := C.ct_callback(context.ctx, nil, C.CS_SET, C.CS_SERVERMSG_CB, C.ct_callback_client_message)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_callback failed for client messages")
		}
	}

	if dsn.Prop("cgo-callback-server") == "yes" {
		retval := C.ct_callback(context.ctx, nil, C.CS_SET, C.CS_CLIENTMSG_CB, C.ct_callback_server_message)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_callback failed for server messages")
		}
	}

	return nil
}
