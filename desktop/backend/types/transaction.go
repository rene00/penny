package types

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// DefaultDBTimeLayout is the time layout used in datastore.
const DefaultDBTimeLayout string = "2006-01-02 15:04:05.000"

type TransactionResponse struct {
	Success bool        `json:"msg"`
	Msg     string      `json:"success"`
	Data    Transaction `json:"data"`
}

// NewTransactionResponse returns a default TransactionResponse. All attributes
// of TransactionResponse must be set or else the struct wont be returned
// correctly to the frontend.
func NewTransactionResponse() TransactionResponse {
	t := TransactionResponse{
		Success: false,
		Msg:     "default",
		Data:    Transaction{},
	}
	return t
}

type Transaction struct {
	ID               int64  `json:"id"`
	Memo             string `json:"memo"`
	Date             string `json:"date"`
	Currency         `json:"currency"`
	Splits           []Split `json:"splits"`
	Amount           string  `json:"amount"`
	modelTransaction *model.Transaction
	logger           *logwrap.LogWrap

	// Display attributes are potentially shown in UI and not used in any other logic.
	Display TransactionDisplay `json:"display"`
}

type TransactionDisplay struct {
	Account TransactionDisplayAccount `json:"account"`
}

type TransactionDisplayAccount struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

func NewTransaction(logger *logwrap.LogWrap, t *model.Transaction) (Transaction, error) {
	transaction := Transaction{
		ID:               t.ID,
		Memo:             t.Memo,
		modelTransaction: t,
		logger:           logger,
	}

	date, err := time.Parse(DefaultDBTimeLayout, t.Date)
	if err != nil {
		return transaction, err
	}
	transaction.Date = date.Format("2006-01-02")

	return transaction, nil
}

func (t *Transaction) SetSplits(ctx context.Context, db *sql.DB) error {
	var q []qm.QueryMod

	q = []qm.QueryMod{
		qm.Select("splits.id, splits.transactions_id, splits.account_id, splits.value_num, splits.value_denom"),
	}
	splits, err := t.modelTransaction.Splits(q...).All(ctx, db)
	if err != nil {
		t.logger.Error(err.Error())
		return err
	}

	s := []Split{}

	for _, ii := range splits {
		account, err := model.FindAccount(ctx, db, ii.AccountID)
		if err != nil {
			t.logger.Error(err.Error())
			return err
		}

		accountType, err := model.FindAccountType(ctx, db, account.AccountTypeID)
		if err != nil {
			t.logger.Error(err.Error())
			return err
		}

		parentAccountType := &ParentAccountType{}
		if accountType.ParentID != 0 {
			pat, err := model.FindAccountType(ctx, db, accountType.ParentID)
			if err != nil {
				t.logger.Error(err.Error())
				return err
			}
			parentAccountType = &ParentAccountType{
				ID:   pat.ID,
				Name: pat.Name,
			}
		}

		split := Split{
			ID:         ii.ID,
			ValueNum:   ii.ValueNum,
			ValueDenom: ii.ValueDenom,
			Amount:     getAmountAsString(ii.ValueNum, ii.ValueDenom),
			Account: Account{
				ID:   account.ID,
				Name: account.Name,
				AccountType: AccountType{
					ID:                accountType.ID,
					Name:              accountType.Name,
					ParentAccountType: parentAccountType,
				},
			},
		}

		s = append(s, split)
	}

	t.Splits = s

	return nil
}

func (t *Transaction) SetDisplay() {
	display := TransactionDisplay{
		Account: TransactionDisplayAccount{},
	}
	switch len(t.Splits) {
	case 1:
	default:
		split := t.Splits[1]
		display = TransactionDisplay{
			Account: TransactionDisplayAccount{
				Name: split.Account.Name,
				ID:   split.Account.ID,
			},
		}
	}
	t.Display = display
	return
}

func (t *Transaction) SetAmount() {
	if len(t.Splits) < 1 {
		t.Amount = "0"
		return
	}

	// TODO: work out how to show proper amount. For now I will just display
	// the amount of the first split.
	firstSplit := t.Splits[0]

	t.Amount = getAmountAsString(firstSplit.ValueNum, firstSplit.ValueDenom)
	if firstSplit.Account.AccountType.ParentAccountType.Name == "Expenses" {
		t.Amount = fmt.Sprintf("-%s", t.Amount)
	}
}

// getAmountAsString accepts a numerator, a denominator and returns the decimal
// amount as a string.
func getAmountAsString(num, denom int64) string {
	r := big.NewRat(num, denom)
	f, _ := r.Float64()
	amount := decimal.NewFromFloat(f)
	return fmt.Sprintf("%s", amount)
}
