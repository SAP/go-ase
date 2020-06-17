package main

import (
	"context"
	"database/sql/driver"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/types"
)

var (
	fMaxColPrintLength = flag.Int("maxColLength", 50, "Maximum number of characters to print for column")
)

func process(conn *cgo.Connection, query string) error {
	cmd, err := conn.GenericExec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("Query failed: %w", err)
	}
	defer cmd.Drop()

	for {
		rows, result, _, err := cmd.Response()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			cmd.Cancel()
			return fmt.Errorf("Reading response failed: %w", err)
		}

		if rows != nil {
			err = processRows(rows)
			if err != nil {
				log.Printf("Error processing rows: %v", err)
			}
		}

		if result != nil {
			err = processResult(result)
			if err != nil {
				log.Printf("error processing result: %v", err)
			}
		}
	}
}

func processRows(rows *cgo.Rows) error {
	colNames := rows.Columns()

	colLengths := map[int]int{}

	fmt.Printf("|")
	for i, colName := range colNames {
		cellLen := int(rows.ColumnTypeMaxLength(i))
		if cellLen > *fMaxColPrintLength {
			cellLen = *fMaxColPrintLength
		}
		s := " %-" + strconv.Itoa(cellLen) + "s |"
		fmt.Printf(s, colName)
		colLengths[i] = cellLen
	}
	fmt.Printf("\n")

	cells := make([]driver.Value, len(colNames))
	cellsI := make([]interface{}, len(colNames))

	for i, cell := range cells {
		cellsI[i] = &cell
	}

	for {
		err := rows.Next(cells)
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("Scanning cells failed: %w", err)
		}

		fmt.Printf("|")
		for i, cell := range cells {
			s := " %-" + strconv.Itoa(colLengths[i]) + "v |"
			switch rows.ColumnTypeDatabaseTypeName(i) {
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

func processResult(result *cgo.Result) error {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Retrieving the affected rows failed: %w", err)
	}

	if affectedRows >= 0 {
		fmt.Printf("Rows affected: %d\n", affectedRows)
	}
	return nil
}
