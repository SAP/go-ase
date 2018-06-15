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
	"database/sql/driver"
	"errors"
	"fmt"
	"unsafe"
)

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

type drv struct{}

var (
	cContext *C.CS_CONTEXT
)

//database connection
type connection struct {
	conn *C.CS_CONNECTION
}

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

func (d *drv) Open(dsn string) (driver.Conn, error) {
	dsnInfo, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// create connection
	var cConnection *C.CS_CONNECTION
	rc := C.ct_con_alloc(cContext, &cConnection)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_con_alloc failed")
	}

	// set user name
	cUsername := unsafe.Pointer(&dsnInfo.Username)
	defer C.free(unsafe.Pointer(cUsername))
	rc = C.ct_con_props(cConnection, C.CS_SET, C.CS_USERNAME, cUsername, C.CS_NULLTERM, nil)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_con_props failed for C.CS_USERNAME")
	}

	// set password
	cPassword := unsafe.Pointer(&dsnInfo.Password)
	defer C.free(unsafe.Pointer(cPassword))
	rc = C.ct_con_props(cConnection, C.CS_SET, C.CS_PASSWORD, cPassword, C.CS_NULLTERM, nil)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_con_props failed for C.CS_PASSWORD")
	}

	// connect
	cHostname := C.CString(dsnInfo.Host)
	cNullterm := (C.long)(C.CS_NULLTERM)
	if dsnInfo.Host != "" {
		cNullterm = (C.long)(0)
	}
	rc = C.ct_connect(cConnection, cHostname, cNullterm)
	if rc != C.CS_SUCCEED {
		return nil, errors.New("C.ct_connect failed")
	}

	// return connection
	return &connection{conn: cConnection}, nil
}

func (connection *connection) Begin() (driver.Tx, error) {
	return &transaction{conn: connection.conn}, nil
}
