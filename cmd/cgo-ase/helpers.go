package main

import (
	"database/sql"
	"fmt"
)

func processRows(rows *sql.Rows) error {
	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("Failed to retrieve column names: %v", err)
	}

	fmt.Printf("|")
	for _, colName := range colNames {
		fmt.Printf(" %s |", colName)
	}
	fmt.Printf("\n")

	cells := make([]interface{}, len(colNames))

	cellsRef := make([]interface{}, len(colNames))
	for i := range cells {
		cellsRef[i] = &(cells[i])
	}

	for rows.Next() {
		err := rows.Scan(cellsRef...)
		if err != nil {
			return fmt.Errorf("Error retrieving rows: %v", err)
		}

		for _, cell := range cells {
			fmt.Printf("| %v ", cell)
		}
		fmt.Printf("|\n")
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("Error preparing rows: %v", err)
	}

	return nil
}

func processResult(result sql.Result) error {
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Retrieving the affected rows failed: %v", err)
	}

	fmt.Printf("Rows affected: %d\n", affectedRows)
	return nil
}
