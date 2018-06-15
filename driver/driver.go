package driver

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -Wl,-rpath,\$ORIGIN
#include <stdlib.h>
#include "ctpublic.h"
*/
import "C"
import (
	"database/sql"
	"fmt"
	"unsafe"
)

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

type drv struct{}

var (
	cContext *C.CS_CONTEXT
)

func init() {
	sql.Register(DriverName, &drv{})
	rc := C.cs_ctx_alloc(C.CS_CURRENT_VERSION, &cContext)
	if rc != C.CS_SUCCEED {
		fmt.Println("C.cs_ctx_alloc failed")
		return
	}
	defer C.free(unsafe.Pointer(cContext))

	rc = C.ct_init(cContext, C.CS_CURRENT_VERSION)
	if rc != C.CS_SUCCEED {
		fmt.Println("C.ct_init failed")
		C.cs_ctx_drop(cContext)
		return
	}
}
