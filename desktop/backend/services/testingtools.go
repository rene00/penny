package services

import (
	"context"
	"fmt"
	"os"
	"pennyapp/backend/config"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"testing"
)

type TestServices struct {
	DB           *dbService
	Accounts     *accountService
	Transactions *transactionService
	Splits       *splitsService
}

func TestSetupApp(t *testing.T) (TestServices, error) {
	t.Helper()

	var err error
	var services TestServices

	ctx := context.Background()
	logger := logwrap.Get("logger")
	if logger == nil {
		logger = logwrap.New("logger", os.Stdout, true)
	}
	logger.SetLevel(logwrap.INFO)

	dbFile, err := os.CreateTemp("/tmp", "pennyapp")
	if err != nil {
		return services, err
	}

	conf := config.NewConfig(config.WithDBFile(dbFile.Name()))
	dbSvc := DB()
	transactionSvc := Transaction()
	splitSvc := Splits()
	accountSvc := Account()
	dbSvc.Start(ctx, conf)
	transactionSvc.Start(ctx, dbSvc.DB, logger)
	splitSvc.Start(ctx, dbSvc.DB, logger)
	accountSvc.Start(ctx, dbSvc.DB, logger)

	services = TestServices{
		DB:           dbSvc,
		Accounts:     accountSvc,
		Transactions: transactionSvc,
		Splits:       splitSvc,
	}

	return services, nil
}

func TestCreateTestTransactions(t *testing.T, s TestServices) (model.Transaction, error) {
	var err error
	var transaction model.Transaction

	transaction, err = s.Transactions.createTransaction("testTransactionMemo", "2024-05-22T20:36:00.000Z")
	if err != nil {
		return transaction, err
	}

	expenseChildAccountType, err := s.Accounts.getChildAccountType("expense")
	if err != nil {
		return transaction, err
	}

	account, err := s.Accounts.createAccount("testAccountName", "testAccountDescription", expenseChildAccountType.ID)
	if err != nil {
		return transaction, err
	}

	splitType := s.Splits.CreateSplit(transaction.ID, account.ID, float64(0))
	if !splitType.Success {
		return transaction, fmt.Errorf("failed create split")
	}

	return transaction, nil
}
