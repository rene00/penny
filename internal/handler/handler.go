package handler

import (
	"github.com/mattn/go-sqlite3"
)

type DB struct {
}

func NewDBHandler() DB {
	return DB{}
}

func (d DB) IsSQLiteErrConstraintUnique(err error) bool {
	return d.isSQLiteError(err, []sqlite3.ErrNoExtended{sqlite3.ErrConstraintUnique})
}

// IsSQLiteError accepts an error and slice of sqlite3.ErrNoExtended and
// returns true if error is in the slice.
func (d DB) isSQLiteError(err error, sqliteErrors []sqlite3.ErrNoExtended) bool {
	if err, ok := err.(sqlite3.Error); ok {
		for _, i := range sqliteErrors {
			if err.ExtendedCode == i {
				return true
			}
		}
	}
	return false
}
