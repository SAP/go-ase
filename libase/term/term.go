// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package term

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"database/sql"
)

var (
	fInputFile = flag.String("f", "", "Read SQL commands from file")
)

func Entrypoint(db *sql.DB) error {
	flag.Parse()

	if len(flag.Args()) == 0 && *fInputFile == "" {
		return Repl(db)
	}

	query := strings.Join(flag.Args(), " ") + ";"

	if *fInputFile != "" {
		bs, err := ioutil.ReadFile(*fInputFile)
		if err != nil {
			return fmt.Errorf("term: error reading file '%s': %w", *fInputFile, err)
		}
		query = string(bs)
	}

	return ParseAndExecQueries(db, query)
}
