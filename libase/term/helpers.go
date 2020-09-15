// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package term

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/SAP/go-ase/libase/types"
)

var (
	fMaxColPrintLength = flag.Int("maxColLength", 50, "Maximum number of characters to print for column")
)

type GenericExecer interface {
	GenericExec(context.Context, string, []driver.NamedValue) (driver.Rows, driver.Result, error)
}

func process(db *sql.DB, query string) error {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error getting sql.Conn: %w", err)
	}
	defer conn.Close()

	return conn.Raw(func(driverConn interface{}) error {
		return rawProcess(driverConn, query)
	})
}

func rawProcess(driverConn interface{}, query string) error {
	execer, ok := driverConn.(GenericExecer)
	if !ok {
		return fmt.Errorf("invalid driver, must support GenericExecer")
	}

	rows, result, err := execer.GenericExec(context.Background(), query, nil)
	if err != nil {
		return fmt.Errorf("GenericExec failed: %w", err)
	}

	if rows != nil && !reflect.ValueOf(rows).IsNil() {
		defer rows.Close()

		if err := processRows(rows); err != nil {
			return fmt.Errorf("error processing rows: %w", err)
		}
	}

	if result != nil && !reflect.ValueOf(result).IsNil() {
		if err := processResult(result); err != nil {
			return fmt.Errorf("error processing result: %w", err)
		}
	}

	return nil
}

func processRows(rows driver.Rows) error {
	rowsColumnTypeLength, ok := rows.(driver.RowsColumnTypeLength)
	if !ok {
		return errors.New("rows does not support driver.RowsColumnTypLength")
	}

	rowsColumnTypeName, ok := rows.(driver.RowsColumnTypeDatabaseTypeName)
	if !ok {
		return errors.New("rows does not support driver.RowsColumnTypesDatabaseTypeName")
	}

	colNames := rows.Columns()
	colLengths := map[int]int{}

	fmt.Printf("|")
	for i, colName := range colNames {
		cellLen := len(colName)

		colTypeLen, ok := rowsColumnTypeLength.ColumnTypeLength(i)
		if ok {
			cellLen = int(colTypeLen)
		}

		if cellLen > *fMaxColPrintLength {
			cellLen = *fMaxColPrintLength
		}
		s := " %-" + strconv.Itoa(cellLen) + "s |"
		fmt.Printf(s, colName)
		colLengths[i] = cellLen
	}
	fmt.Printf("\n")

	cells := make([]driver.Value, len(colNames))

	for {
		if err := rows.Next(cells); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("scanning cells failed: %w", err)
		}

		fmt.Printf("|")
		for i, cell := range cells {
			s := " %-" + strconv.Itoa(colLengths[i]) + "v |"
			switch rowsColumnTypeName.ColumnTypeDatabaseTypeName(i) {
			case "DECIMAL":
				fmt.Printf(s, cell.(*types.Decimal).String())
			case "IMAGE":
				b := hex.EncodeToString(cell.([]byte))
				if len(b) > colLengths[i] {
					fmt.Printf(s, b[:colLengths[i]])
				} else {
					fmt.Printf(s, b)
				}
			default:
				fmt.Printf(s, (interface{})(cell))
			}
		}
		fmt.Printf("\n")
	}

	return nil
}

func processResult(result sql.Result) error {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Retrieving the affected rows failed: %w", err)
	}

	if affectedRows >= 0 {
		fmt.Printf("Rows affected: %d\n", affectedRows)
	}
	return nil
}
