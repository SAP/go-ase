# go-ase

`go-ase` provides a cgo driver for `database/sql`.

A pure go driver is planned.

## Supported ASE data types

| ASE data type | Golang data type  |
| ------------- | ----------------- |
| BIGINT        | int64             |
| FLOAT         | float64           |
| BIT           | bool              |
| BINARY        | []byte            |
| CHAR          | string            |
| BIGDATETIME   | time.Time         |

## cgo
The `cgo` driver is a shim for Client-Library and requires the shared
objects from Client-Library for compiling. The headers are provided at
`cgo/includes`.

The Client-Libray SDK installation guide is available here:

[https://help.sap.com/viewer/882ef48c7e9c4d6e845d98f34378db40/16.0.3.2/en-US]()

### Usage

Example code:

```go
package main

import (
    _ "github.com/SAP/go-ase/cgo"
)

func main() {
    db, err := sql.Open("ase", "ase://user:pass@host:port/")
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Printf("Failed to ping database: %v", err)
        return
    }
}
```

Compilation:

```sh
CGO_LDFLAGS="-L/path/to/OCS/lib -lsybct_r64 -lsybcs_r64" go build -o cmd ./
```

Execution:

```sh
LD_LIBRARY_PATH="/path/to/OCS/lib" ./cmd
```

`/path/to/OCS/lib` is the path to your Client-Library installation's
shared libraries.

## Tests

### Unit tests

Unit tests for the packages are included in their respective directories
and can be run using `go test`.

### Integration tests

Integration tests are available in `tests/` and can be run using `go
test test/${type}test`, where `$type` is either `go` or `cgo`.

These require the following environment variables to be set:

- `ASE_HOST`
- `ASE_PORT`
- `ASE_USER`
- `ASE_PASS`

The cgo tests additionally require the variable `ASE_USERSTOREKEY` to be
set.
