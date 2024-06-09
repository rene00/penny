package services

import (
	"os"
	"testing"
)

func TestGetTransaction(t *testing.T) {
	t.Run("standard get transaction", func(t *testing.T) {
		s, err := TestSetupApp(t)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(s.DB.dbFile)

		transaction, err := TestCreateTestTransactions(t, s)
		if err != nil {
			t.Fatal(err)
		}

		transactionResponse := s.Transactions.GetTransaction(transaction.ID)
		if !transactionResponse.Success {
			t.Fatal()
		}

		if transactionResponse.Msg != "default" {
			t.Fatal()
		}

		if transactionResponse.Data.Memo != "testTransactionMemo" {
			t.Fatal(transactionResponse.Data.Memo)
		}

		if len(transactionResponse.Data.Splits) == 0 {
			t.Fatal()
		}
	})
}

func TestSplits(t *testing.T) {
	var err error

	s, err := TestSetupApp(t)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(s.DB.dbFile)

	transaction, err := s.Transactions.createTransaction("foo", "2024-05-22T20:36:00.000Z")
	if err != nil {
		t.Fatal(err)
	}

	expenseChildAccountType, err := s.Accounts.getChildAccountType("expense")
	if err != nil {
		t.Fatal(err)
	}

	account, err := s.Accounts.createAccount("test", "test", expenseChildAccountType.ID)
	if err != nil {
		t.Fatal(err)
	}

	splitType := s.Splits.CreateSplit(transaction.ID, account.ID, float64(0))
	if !splitType.Success {
		t.Fatal()
	}

	transactionResponse := s.Transactions.GetTransaction(transaction.ID)
	if transactionResponse.Data.Memo != "foo" {
		t.Fatal()
	}

	if transactionResponse.Data.Splits[0].Account.Name != "test" {
		t.Fatal()
	}
}
