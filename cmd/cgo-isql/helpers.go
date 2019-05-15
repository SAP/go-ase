package main

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"

	ase "github.com/SAP/go-ase/cgo"
)

func processRows(rows *ase.Rows) error {
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
			fmt.Printf(s, (interface{})(cell))
		}
		fmt.Printf("|\n")
	}

	return nil
}

func processResult(result *ase.Result) error {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Retrieving the affected rows failed: %v", err)
	}

	if affectedRows >= 0 {
		fmt.Printf("Rows affected: %d\n", affectedRows)
	}
	return nil
}
