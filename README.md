# go-ase

## Description

`go-ase` is an SAP ASE database driver for Go (golang) and its `database/sql` package.
It is delivered as Go module.
- The `cgo` driver is a shim for the Client-Library.
- A pure go driver is planned.

## Requirements

The `cgo` driver requires the shared objects from the Client-Library for compiling.
The headers are provided at `cgo/includes`. For more details please see the
[Client-Libray SDK installation guide][cl-sdk-install-guide].

## Download and Installation

### cgo

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

`/path/to/OCS/lib` is the path to your Client-Library installation's shared libraries.

### Unit tests

Unit tests for the packages are included in their respective directories
and can be run using `go test`.

### Integration tests

Integration tests are available in `tests/` and can be run using `go test test/${type}test`,
where `$type` is either `go` or `cgo`.

These require the following environment variables to be set:

- `ASE_HOST`
- `ASE_PORT`
- `ASE_USER`
- `ASE_PASS`

The cgo tests additionally require the variable `ASE_USERSTOREKEY` to be set.

The integration tests will create new databases for each connection type to run tests
against. After the tests are finished the created databases will be removed.

## Configuration

### Connection Properties

The data source name (DSN) to specify the connection properties can be given as:
- a URI based DSN string,
- a simple DSN string or
- a `dsn.DsnInfo`.

#### URI DSN

The URI DSN is a common URI like `ase://user:pass@host:port/?prop1=val1&prop2=val2`.

DSNs in this form are parsed using `url.Parse`.

#### Simple DSN

The simple DSN is a key/value string: `username=user password=pass host=hostname port=4901`

Values with spaces must be quoted using single or double quotes.

Each member of `libase.dsn.DsnInfo` can be set using any of their
possible json tags. E.g. `.Host` will receive the values from the keys
`host` and `hostname`.

Additional properties are set as key/value pairs as well: `...
prop1=val1 prop2=val2`. If the parser doesn't recognize a string as
a json tag it assumes that the key/value pair is a property and its
value.

Similar to the URI DSN those property/value pairs are purely additive.
Any property that only recognizes a single argument (e.g. a boolean)
will only honour the last given value for a property.

#### Connector DSN

As an alternative to the string DSNs `cgo.NewConnector` accepts
a `dsn.DsnInfo` directly and returns a `driver.Connector`, which can be
passed to `sql.OpenDB`:

```go
package main

import (
    "database/sql"

    "github.com/SAP/go-ase/libase/dsn"
    ase "github.com/SAP/go-ase/cgo"
)

func main() {
    d := dsn.NewDsnInfo()
    d.Host = "hostname"
    d.Port = "4901"
    d.Username = "user"
    d.Password = "pass"

    connector, err := ase.NewConnector(d)
    if err != nil {
        log.Printf("Failed to create connector: %v", err)
        return
    }

    db, err := sql.OpenDB(connector)
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    _, err = db.Exec("select 'ping'")
    if err != nil {
        log.Printf("Failed to ping ASE: %v", err)
        return
    }
}
```

### cgo Properties

#### cgo-callback-client

Recognized values: `yes` or any string

When set to `yes` the callback for client messages is set. By default
these messages are printed to stderr.

When set to any other string the callback will not bet set.

These messages signal a local error in Client-Library.

#### cgo-callback-server

Recognized values: `yes` or any string

When set to `yes` the callback for server messages is set. By default
these messages printed to stderr.

When set to any other string the callback will not be set.

These messages signal an error in the ASE while processing a request.

## Limitations

### Supported ASE data types

| ASE data type | Golang data type  |
| ------------- | ----------------- |
| BIGINT        | int64             |
| FLOAT         | float64           |
| BIT           | bool              |
| BINARY        | []byte            |
| CHAR          | string            |
| BIGDATETIME   | time.Time         |

## Known Issues

The list of known issues is available [here][issues].

## How to obtain support

Feel free to open issues for feature requests, bugs or general feedback [here][issues].

## Contributing

Any help to improve this package is highly appreciated.

## To-Do (upcoming changes)

A pure go driver is planned.

## License

Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
This file is licensed under the Apache License 2.0 except as noted otherwise in the [LICENSE file](LICENSE)

[issues]: https://github.com/SAP/go-ase/issues
[cl-sdk-install-guide]: https://help.sap.com/viewer/882ef48c7e9c4d6e845d98f34378db40/16.0.3.2/en-US
