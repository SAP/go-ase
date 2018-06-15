package driver

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -Wl,-rpath,\$ORIGIN
#include <stdlib.h>
#include "ctpublic.h"
*/
import "C"

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

type drv struct{}
