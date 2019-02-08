package cgo

// #include "ctlib.h"
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// srvMsg is a callback function which will be called from C when the server sends a message.
// This may be an error message or e.g. the information that we are connected successfully.
// Don't change the following line. It is the directive for cgo to make the function available from C.
//export srvMsg
func srvMsg(msg *C.CS_SERVERMSG) C.CS_RETCODE {
	switch msg.msgnumber {
	case C.CS_SV_INFORM:
		break
	default:
		fmt.Fprintln(os.Stderr, "Server message:")
		fmt.Fprintf(os.Stderr, "msgnumber:   %d\n", msg.msgnumber)
		fmt.Fprintf(os.Stderr, "state:       %d\n", msg.state)
		fmt.Fprintf(os.Stderr, "severity:    %d\n", msg.severity)
		fmt.Fprintf(os.Stderr, "text:        %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.text))))
		fmt.Fprintf(os.Stderr, "textlen:     %d\n", msg.textlen)
		fmt.Fprintf(os.Stderr, "server:      %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.svrname))))
		fmt.Fprintf(os.Stderr, "svrnlen:     %d\n", msg.svrnlen)
		fmt.Fprintf(os.Stderr, "proc:        %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.proc))))
		fmt.Fprintf(os.Stderr, "proclen:     %d\n", msg.proclen)
		fmt.Fprintf(os.Stderr, "line:        %d\n", msg.line)
		fmt.Fprintf(os.Stderr, "status:      %d\n", msg.status)
		fmt.Fprintf(os.Stderr, "sqlstate:    %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.sqlstate))))
		fmt.Fprintf(os.Stderr, "sqlstatelen: %d\n", msg.sqlstatelen)
	}

	return C.CS_SUCCEED
}

// ctlMsg is a callback function which will be called from C when the client sends a message.
// This may be an error message or some information.
// Don't change the following line. It is the directive for cgo to make the function available from C.
//export ctlMsg
func ctlMsg(msg *C.CS_CLIENTMSG) C.CS_RETCODE {
	fmt.Fprintln(os.Stderr, "Client message:")
	fmt.Fprintf(os.Stderr, "severity:     %d\n", msg.severity)
	fmt.Fprintf(os.Stderr, "msgnumber:    %d\n", msg.msgnumber)
	fmt.Fprintf(os.Stderr, "msgstring:    %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.msgstring))))
	fmt.Fprintf(os.Stderr, "msgstringlen: %d\n", msg.msgstringlen)
	fmt.Fprintf(os.Stderr, "osnumber:     %d\n", msg.osnumber)
	fmt.Fprintf(os.Stderr, "osstring:     %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.osstring))))
	fmt.Fprintf(os.Stderr, "osstringlen:  %d\n", msg.osstringlen)
	fmt.Fprintf(os.Stderr, "status:       %d\n", msg.status)
	fmt.Fprintf(os.Stderr, "sqlstate:     %s\n", C.GoString((*C.char)(unsafe.Pointer(&msg.sqlstate))))
	fmt.Fprintf(os.Stderr, "sqlstatelen:  %d\n", msg.sqlstatelen)

	return C.CS_SUCCEED
}
