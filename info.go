// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"flag"
	"fmt"

	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/tds"
)

type Info struct {
	tds.Info

	AppName string `json:"appname" doc:"Application Name to transmit to ASE"`

	NoQueryCursor bool `json:"no-query-cursor" doc:"Prevents the use of cursors for database/sql query methods. See README for details."`
}

// NewInfo returns a bare Info for github.com/SAP/go-dblib/dsn with defaults.
func NewInfo() (*Info, error) {
	info := new(Info)

	if err := tds.SetInfo(&info.Info); err != nil {
		return nil, fmt.Errorf("ase: error setting TDS defaults on info: %w", err)
	}

	info.AppName = "github.com/SAP/go-ase"

	return info, nil
}

// NewInfoWithEnv is a convenience function returning an Info with
// values filled from the environment with the prefix 'ASE'.
func NewInfoWithEnv() (*Info, error) {
	info, err := NewInfo()
	if err != nil {
		return nil, err
	}

	if err := dsn.FromEnv("ASE", info); err != nil {
		return nil, fmt.Errorf("ase: error setting environment values on info: %w", err)
	}

	return info, nil
}

// NewInfoFlags is a convenience function returning an Info filled with
// defaults and a flagset with flags bound to the members of the
// returned info.
func NewInfoWithFlags() (*Info, *flag.FlagSet, error) {
	info, err := NewInfo()
	if err != nil {
		return nil, nil, err
	}

	flagset, err := dsn.FlagSet("", flag.ContinueOnError, info)
	if err != nil {
		return nil, nil, fmt.Errorf("ase: error creating flagset: %w", err)
	}

	return info, flagset, nil
}
