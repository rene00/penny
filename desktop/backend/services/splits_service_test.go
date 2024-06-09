package services

import (
	"context"
	"os"
	"testing"
)

func TestGetSplit(t *testing.T) {
	t.Run("test getSplit() returns all relations", func(t *testing.T) {
		ctx := context.Background()
		s, err := TestSetupApp(t)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(s.DB.dbFile)

		transaction, err := TestCreateTestTransactions(t, s)
		if err != nil {
			t.Fatal(err)
		}

		splits, err := transaction.Splits().All(ctx, s.DB.DB)
		if err != nil {
			t.Fatal(err)
		}

		split, err := s.Splits.getSplit(splits[0].ID)
		if err != nil {
			t.Fatal(err)
		}

		a, err := split.Account().One(ctx, s.DB.DB)
		if err != nil {
			t.Fatal(err)
		}
		if a.Name != "testAccountName" {
			t.Fatal(err)
		}

		at, err := a.AccountType().One(ctx, s.DB.DB)
		if err != nil {
			t.Fatal(err)
		}
		if at.Name != "Expense" {
			t.Fatal(at)
		}

	})
}
