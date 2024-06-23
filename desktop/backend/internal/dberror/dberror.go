package dberror

import (
	"database/sql"
	"errors"
	"strings"
)

func IsNoRowsFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func IsUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
