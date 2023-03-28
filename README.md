<!--
SPDX-FileCopyrightText: 2020 SAP SE
SPDX-FileCopyrightText: 2021 SAP SE
SPDX-FileCopyrightText: 2022 SAP SE
SPDX-FileCopyrightText: 2023 SAP SE

SPDX-License-Identifier: Apache-2.0
-->

# go-ase

[![PkgGoDev](https://pkg.go.dev/badge/github.com/SAP/go-ase)](https://pkg.go.dev/github.com/SAP/go-ase)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/go-ase)](https://goreportcard.com/report/github.com/SAP/go-ase)
[![REUSE
status](https://api.reuse.software/badge/github.com/SAP/go-ase)](https://api.reuse.software/info/github.com/SAP/go-ase)
![Actions: CI](https://github.com/SAP/go-ase/workflows/CI/badge.svg)

## Description

`go-ase` is a driver for the [`database/sql`][pkg-database-sql] package
of [Go (golang)][go] to provide access to SAP ASE instances.
It is delivered as Go module.

SAP ASE is the shorthand for [SAP Adaptive Server Enterprise][sap-ase],
a relational model database server originally known as Sybase SQL
Server.

A cgo implementation can be found [here][cgo-ase].

## Requirements

The go driver has no special requirements other than Go standard
library and the third part modules listed in `go.mod`, e.g.
`github.com/SAP/go-dblib`.

## Download and Installation

The packages in this repo can be `go get` and imported as usual, e.g.:

```sh
go get github.com/SAP/go-ase
```

For specifics on how to use `database/sql` please see the
[documentation][pkg-database-sql].

The command-line application `goase` can be `go install`ed:

```sh
$ go install github.com/SAP/go-ase/cmd/goase@latest
go: downloading github.com/SAP/go-ase v0.0.0-20210506093950-9af676a6bab4
$ goase -h
Usage of goase:
      --appname string                   Application Name to transmit to ASE (default "github.com/SAP/go-ase")
      --channel-package-queue-size int   How many TDS packages can be queued in a TDS channel (default 100)
      --client-hostname string           Hostname to send to server (default "dev-ase-sles15sp1-ntnn-1")
      --cursor-cache-rows int            How many rows to cache at once when reading the result set of a cursor (default 1000)
      --database string                  Database
      --debug-log-packages               Log packages as they are transmitted/received
  -f, --f string                         Read SQL commands from file
      --host string                      Hostname to connect to
      --maxColLength int                 Maximum number of characters to print for column (default 50)
      --network string                   Network to use, either 'tcp' or 'udp' (default "tcp")
      --no-query-cursor                  Prevents the use of cursors for database/sql query methods. See README for details.
      --packet-read-timeout int          Time in seconds to wait before aborting a connection when no response is received from the server (default 50)
      --password string                  Password
      --port string                      Port (Example: '443' or 'tls') to connect to
      --tls-ca-file string               Path to CA file to validate server certificate against
      --tls-enable                       Enforce TLS use
      --tls-hostname string              Remote hostname to validate against SANs
      --tls-skip-validation              Skip TLS validation - accepts any TLS certificate
      --username string                  Username
2021/05/06 11:31:44 goase failed: pflag: help requested
```

## Usage

Example code:

```go
package main

import (
    "database/sql"
    _ "github.com/SAP/go-ase"
)

func main() {
    db, err := sql.Open("ase", "ase://user:pass@host:port/")
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Printf("Failed to ping database: %v", err)
        return
    }
}
```

### Compilation

```sh
go build -o goase ./cmd/goase/
```

### Execution

```sh
./goase
```

### Examples

More examples can be found in the folder `examples`.

### Integration tests

Integration tests are available and can be run using `go test --tags=integration` and
`go test ./examples/... --tags=integration`.

These require the following environment variables to be set:

- `ASE_HOST`
- `ASE_PORT`
- `ASE_USER`
- `ASE_PASS`

The integration tests will create new databases for each connection type to run tests
against. After the tests are finished the created databases will be removed.

## Configuration

The configuration is handled through either a data source name (DSN) in
one of two forms or through a configuration struct passed to a connector.

All of these support additional properties which can tweak the
connection or the drivers themselves.

### Data Source Names

#### URI DSN

The URI DSN is a common URI like `ase://user:pass@host:port/?prop1=val1&prop2=val2`.

DSNs in this form are parsed using `url.Parse`.

#### Simple DSN

The simple DSN is a key/value string: `username=user password=pass host=hostname port=4901`

Each member of `Info` is recognized by its `json` metadata tag or any of
its `multiref` metadata tags.

Values with spaces must be quoted using single or double quotes.

#### Connector

As an alternative to the string DSNs `ase.NewConnector` accept a `Info`
directly and return a `driver.Connector`, which can be passed to
`sql.OpenDB`:

```go
package main

import (
    "database/sql"

    "github.com/SAP/go-ase"
)

func main() {
    info := ase.NewInfo()
    info.Host = "hostname"
    info.Port = "4901"
    info.Username = "user"
    info.Password = "pass"

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

    if err := db.Ping(); err != nil {
        log.Printf("Failed to ping ASE: %v", err)
    }
}
```

### Properties

##### appname

Recognized values: string

Sets the application name to the value. This can be used in ASE to
determine which application opened a connection.

Defaults to `database/sql driver github.com/SAP/go-ase/purego`.

##### network

Recognized values: string

The network must be a network type recognized by `net.Dial` - at the
time of writing this is either `udp` or `tcp`.

This should only be required to be set if the database is only reachable
through a UDP proxy.

Defaults to `tcp`.

##### channel-package-queue-size

Recognized values: integer

Defines how many packages a TDS channel can buffer at most. When working
with very large datasets where heavy computation only occurs every
hundred packages or so it may be feasible to improve performance by
increasing the queue size.

Defaults to 100.

##### client-hostname

Recognized values: string

The client-hostname to report to the TDS server. Due to protocol
limitations this will be cut off after 30 characters.

Defaults to the hostname of the machine, acquired using `os.Hostname`.

##### packet-read-timeout

Recognized values: integer

The timeout in seconds when a packet is read. The timeout is reset every
time a packet successfully reads data from the connection.

That means the timeout only triggers if no data was read for longer than
`packet-read-timeout` seconds.

Default to 50.

##### tls

Recognized values: bool

Activates TLS for the connection. Any other TLS option is ignored unless tls is set to true.

Defaults to true if the target port is 443, false otherwise.

##### tls-hostname

Recognized values: string

Allows to pass SAN for TLS validation.

For compatibility with the cgo implementation you may also use `ssl`
instead of `tls-hostname` and pass `CN=<SAN>` instead of `<SAN>`.

Defaults to empty string.

Please note that as of go1.15 the CommonName in x509 certificates is no
longer recognized as the hostname if no SANs are present in the
certificate.
If the certificate for your TDS server only utilizes the CN you can
reenable this behaviour by setting `GODEBUG` to `x509ignoreCN=0` in your
environment:

```sh
GODEBUG=x509ignoreCN=0 <path/to/your/app>
GODEBUG=x509ignoreCN=0 go run ./cmd/goase
```

For details see https://golang.google.cn/doc/go1.15#commonname

##### tls-skip-validation

Recognized values: string

If the value is recognized by `strconv.ParseBool` to represent `true`
the TLS certificate of the TDS server will not be validated.

Defaults to empty string / false.

##### tls-ca

Recognized values: string

Path to a CA file, which may contain multiple CAs, to validate the TDS
servers certificate against.
If empty the servers trust store is used.

Defaults to empty string.

##### no-query-cursor

Recognized values: bool

Prevents the use of cursors for database/sql query methods.

By default go-ase creates cursors when a `.Query*` method is used.

Depending on the query, expected result set, memory consumption and
used hardware this may perform worse than a simple query.

It is strongly suggested to profile this option with your queries before
enabling it.

Altneratively you can pass this option on a per-query basis in the context.
See the documentation of Conn.QueryContext for details.

### Nullable data types

Nullable data types are implemented in [go-dblib][go-dblib]. However,
the implementation differs partially from drivers like `isql` in regard
of "zero-length non-Null" string-types, e.g. `""`. Instead of inserting
such values as `" "`, the methods `stmt.Exec(...)` or `db.Exec(...)`
will insert these values as actual `NULL` values. A `legacy`-option to
insert such "zero-length non-Null" string-type values as `" "` is
planned but not implemented yet. In the meantime, it is possible to
reproduce this behaviour by using language tokens as provided by the
executable `goase`.

## Limitations

### Beta

The go implementation is currently in beta and under active development.
As such most features of the TDS protocol and ASE are not supported.

### Prepared statements

Regarding the limitations of prepared statements/dynamic SQL please see
[the Client-Library documentation](https://help.sap.com/viewer/71b47f4a8269411da6d15ed25f5d39b3/LATEST/en-US/bfc531e46db61014bf8f040071e613d7.html).

The Client-Library documentation applies to the go implementation as
these restrictions are imposed by the implementation of dynamic SQL
on the server side.

### Unsupported ASE data types

Currently the following data types are not supported:

- Timestamp
- Univarchar
- Nullable Text
- Nullable Unitext
- Nullable Unichar
- Nullable Image

## Known Issues

The list of known issues is available [here][issues].

## How to obtain support

Feel free to open issues for feature requests, bugs or general feedback [here][issues].

## Contributing

Any help to improve this package is highly appreciated.

For details on how to contribute please see the
[contributing](CONTRIBUTING.md) file.

## License

Copyright (c) 2019-2020 SAP SE or an SAP affiliate company. All rights reserved.
This file is licensed under the Apache License 2.0 except as noted otherwise in the [LICENSE file](LICENSES).

[cgo-ase]: https://github.com/SAP/cgo-ase
[go-dblib]: https://github.com/SAP/go-dblib
[go]: https://golang.org/
[issues]: https://github.com/SAP/go-ase/issues
[pkg-database-sql]: https://golang.org/pkg/database/sql
[sap-ase]: https://www.sap.com/products/sybase-ase.html
