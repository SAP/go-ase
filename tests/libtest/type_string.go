package libtest

import (
	"database/sql"
	"strings"
	"testing"
)

// DoTestString tests that the string type is handled correctly.
func DoTestString(t *testing.T) {
	TestForEachDB("TestString", t, testString)
}

func testString(t *testing.T, db *sql.DB, tableName string) {
	samples := []string{
		"test",
		"a longer test",
	}

	// 255 is C.CS_MAX_CHAR but can't be referenced as such since go
	// doesn't like using cgo in tests.
	stringLengths := []int{0, 50, 150, 255}

	for _, length := range stringLengths {
		chrs := make([]rune, length)
		chr := 'a'
		for i := 0; i < length; i++ {
			chrs[i] = chr
			chr++
			if chr > 'z' {
				chr -= 26
			}
		}
		samples = append(samples, string(chrs))
	}

	pass := make([]interface{}, len(samples))
	for i, sample := range samples {
		pass[i] = sample
	}
	rows, err := SetupTableInsert(db, tableName, "char(255)", pass...)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer rows.Close()

	i := 0
	var recv string
	for rows.Next() {
		err = rows.Scan(&recv)
		if err != nil {
			t.Errorf("Scan failed on %dth scan: %v", i, err)
			continue
		}

		recv = strings.TrimSpace(recv)
		if strings.Compare(recv, samples[i]) != 0 {
			t.Errorf("Received value does not match passed parameter")
			t.Errorf("Expected: %v", samples[i])
			t.Errorf("Received: %v", recv)
		}

		i += 1
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Error preparing rows: %v", err)
	}
}
