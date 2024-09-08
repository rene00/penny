package types

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/model"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// DefaultDBTimeLayout is the time layout used in datastore.
const DefaultDBTimeLayout string = "2006-01-02 15:04:05.000"

type TransactionResponse struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Data    Transaction `json:"data"`
}

type TransactionsResponse struct {
	Success bool          `json:"success"`
	Msg     string        `json:"msg"`
	Data    []Transaction `json:"data"`
}

// TransactionsTotalBalance is the total balance amount for all non-deleted
// transactions.
type TransactionsTotalBalance struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    string `json:"data"`
}

// TransactionsNetAssetsResponse is the total net assets amount for all
// non-deleted transactions.
type TransactionsNetAssetsResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    string `json:"data"`
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
	ID       int64              `json:"id"`
	Memo     string             `json:"memo"`
	Date     string             `json:"date"`
	Splits   []Split            `json:"splits"`
	Amount   string             `json:"amount"`
	Deleted  bool               `json:"deleted"`
	Display  TransactionDisplay `json:"display"`
	Currency `json:"currency"`
}

type TransactionDisplay struct {
	Account TransactionDisplayAccount `json:"account"`
}

type TransactionDisplayAccount struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

func NewTransaction(ctx context.Context, db *sql.DB, t *model.Transaction) (Transaction, error) {
	var err error

	transaction := Transaction{
		ID:   t.ID,
		Memo: t.Memo,
	}

	date, err := time.Parse(DefaultDBTimeLayout, t.Date)
	if err != nil {
		return transaction, err
	}
	transaction.Date = date.Format("2006-01-02")

	if err = transaction.setSplits(ctx, db, t); err != nil {
		return transaction, err
	}

	if err = transaction.setDeleted(ctx, db, t); err != nil {
		return transaction, err
	}

	transaction.setAmount()
	transaction.setDisplay()

	return transaction, nil
}

func (t *Transaction) setSplits(ctx context.Context, db *sql.DB, transaction *model.Transaction) error {
	splits, err := transaction.Splits(qm.Select("splits.id, splits.transactions_id, splits.account_id, splits.value_num, splits.value_denom")).All(ctx, db)
	if err != nil {
		return err
	}

	s := []Split{}

	for _, ii := range splits {
		account, err := model.FindAccount(ctx, db, ii.AccountID)
		if err != nil {
			return err
		}

		accountType, err := model.FindAccountType(ctx, db, account.AccountTypeID)
		if err != nil {
			return err
		}

		parentAccountType := &ParentAccountType{}
		if accountType.ParentID != 0 {
			pat, err := model.FindAccountType(ctx, db, accountType.ParentID)
			if err != nil {
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

func (t *Transaction) setDisplay() {
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

func (t *Transaction) setAmount() {
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

func (t *Transaction) setDeleted(ctx context.Context, db *sql.DB, transaction *model.Transaction) error {
	transactionAttribute, err := model.TransactionsAttributes(qm.Where("transactions_id=? AND name=?", transaction.ID, "deleted")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		if transactionAttribute.Value == "true" {
			t.Deleted = true
		}
	}
	return nil
}

// getAmountAsString accepts a numerator, a denominator and returns the decimal
// amount as a string.
func getAmountAsString(num, denom int64) string {
	r := big.NewRat(num, denom)
	f, _ := r.Float64()
	return decimal.NewFromFloat(f).StringFixed(2)
}
