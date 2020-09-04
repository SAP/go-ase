// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package term

import (
	"flag"
	"fmt"
	"strings"

	"github.com/SAP/go-ase/libase/flagslice"
	"github.com/SAP/go-ase/libase/libdsn"
)

var (
	fHost         = flag.String("H", "", "database hostname")
	fPort         = flag.String("P", "", "database sql port")
	fUser         = flag.String("u", "", "database user name")
	fPass         = flag.String("p", "", "database user password")
	fUserstorekey = flag.String("k", "", "userstorekey")
	fDatabase     = flag.String("D", "", "database")

	fOpts = &flagslice.FlagStringSlice{}
)

func init() {
	flag.Var(fOpts, "o", "Connection properties")
	flag.Parse()
}

func Dsn() (*libdsn.Info, error) {
	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return nil, fmt.Errorf("error reading DSN info from env: %w", err)
	}

	if *fHost != "" {
		dsn.Host = *fHost
	}

	if *fPort != "" {
		dsn.Port = *fPort
	}

	if *fUser != "" {
		dsn.Username = *fUser
	}

	if *fPass != "" {
		dsn.Password = *fPass
	}

	if *fUserstorekey != "" {
		dsn.Userstorekey = *fUserstorekey
	}

	if *fDatabase != "" {
		dsn.Database = *fDatabase
	}

	for _, fOpt := range fOpts.Slice() {
		split := strings.SplitN(fOpt, "=", 2)
		opt := split[0]
		value := ""
		if len(split) > 1 {
			value = split[1]
		}

		if err := dsn.SetField(opt, value); err != nil {
			return nil, fmt.Errorf("term: error setting field %s with value '%v': %w",
				opt, value, err)
		}
	}

	return dsn, nil
}
