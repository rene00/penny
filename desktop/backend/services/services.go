package services

import (
	"time"
)

const sqliteTimeFormat string = "2006-01-02 03:04:05.000"

func GetSQLiteDateFromVueJSDate(s string) (string, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", err
	}
	return t.Format(sqliteTimeFormat), nil
}

func GetVUSJSdateFromSQLiteDate(s string) (string, error) {
	t, err := time.Parse(sqliteTimeFormat, s)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}
