package migrations

import (
	"testing"
)

func TestDoMigrateDb(t *testing.T) {
	if err := DoMigrateDb("sqlite3://:memory:"); err != nil {
		t.Error(err)
	}
}
