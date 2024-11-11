package types

import (
	"context"
	"database/sql"
	"math/big"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/model"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AccountResponse struct {
	Success bool    `json:"success"`
	Msg     string  `json:"msg"`
	Data    Account `json:"data"`
}

type Account struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	AccountType AccountType `json:"account_type"`
	Amount      string      `json:"amount"`
	Deleted     bool        `json:"deleted"`
	AccountOpts
}

type AccountType struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	*ParentAccountType `json:"parent_account_type"`
}

type ParentAccountType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewAccountResponse() AccountResponse {
	a := AccountResponse{
		Success: false,
		Msg:     "default",
		Data:    Account{},
	}
	return a
}

type AccountOpts struct {
	amountAlwaysPositive bool
}

type AccountOptFunc func(*AccountOpts)

func WithAccountOptsAmountAlwaysPosiive(b bool) AccountOptFunc {
	return func(opts *AccountOpts) {
		opts.amountAlwaysPositive = b
	}
}

func defaultAccountOpts() AccountOpts {
	return AccountOpts{
		amountAlwaysPositive: false,
	}
}

func NewAccount(ctx context.Context, db *sql.DB, a *model.Account, opts ...AccountOptFunc) (Account, error) {
	var err error

	o := defaultAccountOpts()
	for _, fn := range opts {
		fn(&o)
	}

	account := Account{
		ID:          a.ID,
		Name:        a.Name,
		AccountOpts: o,
	}

	if a.R != nil && a.R.AccountType != nil {
		account.AccountType = AccountType{
			ID:   a.R.AccountType.ID,
			Name: a.R.AccountType.Name,
		}
	}

	if err = account.setAmount(ctx, db, a); err != nil {
		return account, err
	}
	if err = account.setDeleted(ctx, db, a); err != nil {
		return account, err
	}
	if err = account.setDescription(ctx, db, a); err != nil {
		return account, err
	}

	return account, err
}

func (a *Account) setAmount(ctx context.Context, db *sql.DB, account *model.Account) error {
	splits, err := model.Splits(qm.Where("account_id=?", account.ID)).All(ctx, db)
	if err != nil {
		return err
	}

	accountType, err := model.AccountTypes(qm.Where("id=?", account.AccountTypeID)).One(ctx, db)
	if err != nil {
		return err
	}

	amount := float64(0)
	for _, split := range splits {
		transactionDeleted, err := isTransactionDeleted(ctx, db, split.TransactionsID)
		if err != nil {
			return err
		}
		if transactionDeleted {
			continue
		}

		r := big.NewRat(split.ValueNum, split.ValueDenom)
		f, _ := r.Float64()

		if !a.amountAlwaysPositive {
			switch strings.ToLower(accountType.Name) {
			case "liability", "income":
				f = -f
			}
		}

		amount = amount + f
	}

	a.Amount = decimal.NewFromFloat(amount).StringFixed(2)
	return nil
}

func (a *Account) setDeleted(ctx context.Context, db *sql.DB, account *model.Account) error {
	accountAttribute, err := model.AccountAttributes(qm.Where("account_id=? AND name=?", account.ID, "deleted")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		if accountAttribute.Value == "true" {
			a.Deleted = true
		}
	}
	return nil
}

func (a *Account) setDescription(ctx context.Context, db *sql.DB, account *model.Account) error {
	accountAttribute, err := model.AccountAttributes(qm.Where("account_id=? AND name=?", account.ID, "description")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		a.Description = accountAttribute.Value
	}
	return nil
}
