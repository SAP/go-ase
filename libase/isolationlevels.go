package libase

import (
	"database/sql"
	"fmt"
)

// IsolationLevel reflects the ASE isolation levels.
type IsolationLevel int

// Valid ASE isolation levels.
const (
	LevelInvalid         IsolationLevel = -1
	LevelReadUncommitted IsolationLevel = iota
	LevelReadCommitted
	LevelRepeatableRead
	LevelSerializableRead
)

var (
	// sql2ase maps sql.IsolationLevel to libase.IsolationLevel.
	sql2ase = map[sql.IsolationLevel]IsolationLevel{
		sql.LevelDefault:         LevelReadCommitted,
		sql.LevelReadUncommitted: LevelReadUncommitted,
		sql.LevelReadCommitted:   LevelReadCommitted,
		sql.LevelWriteCommitted:  LevelInvalid,
		sql.LevelRepeatableRead:  LevelRepeatableRead,
		sql.LevelSerializable:    LevelSerializableRead,
		sql.LevelLinearizable:    LevelInvalid,
	}
)

// IsolationLevelFromGo take a database/sql.IsolationLevel and returns
// the relevant isolation level for ASE.
func IsolationLevelFromGo(lvl sql.IsolationLevel) (IsolationLevel, error) {
	aseLvl, ok := sql2ase[lvl]
	if !ok {
		return LevelInvalid, fmt.Errorf("Unknown database/sql.IsolationLevel: %v", lvl)
	}

	if aseLvl == LevelInvalid {
		return LevelInvalid, fmt.Errorf("Isolation level %v is not supported by ASE")
	}

	return aseLvl, nil
}

// ToGo returns the database/sql.IsolationLevel equivalent of the ASE
// isolation level.
func (lvl IsolationLevel) ToGo() sql.IsolationLevel {
	for sqlLvl, aseLvl := range sql2ase {
		if aseLvl == lvl {
			return sqlLvl
		}
	}

	return sql.LevelDefault
}

// String implements the Stringer interface.
func (lvl IsolationLevel) String() string {
	return lvl.ToGo().String()
}
