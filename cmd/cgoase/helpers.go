package main

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/types"
)

func process(conn *cgo.Connection, query string) error {
	cmd, err := conn.GenericExec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("Query failed: %v", err)
	}
	defer cmd.Drop()

	for {
		rows, result, err := cmd.Response()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			cmd.Cancel()
			return fmt.Errorf("Reading response failed: %v", err)
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

	fmt.Printf("|")
	for i, colName := range colNames {
		s := " %-" + strconv.Itoa(int(rows.ColumnTypeMaxLength(i))) + "s |"
		fmt.Printf(s, colName)
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

			return fmt.Errorf("Scanning cells failed: %v", err)
		}

		for i, cell := range cells {
			s := "| %-" + strconv.Itoa(int(rows.ColumnTypeMaxLength(i))) + "v "
			switch cell.(type) {
			case *types.Decimal:
				fmt.Printf(s, cell.(*types.Decimal).String())
			default:
				fmt.Printf(s, (interface{})(cell))
			}
		}
		fmt.Printf("|\n")
	}

	return nil
}

func processResult(result *cgo.Result) error {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Retrieving the affected rows failed: %v", err)
	}

	if affectedRows >= 0 {
		fmt.Printf("Rows affected: %d\n", affectedRows)
	}
	return nil
}
