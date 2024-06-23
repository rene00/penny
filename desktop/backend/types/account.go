package types

import (
	"context"
	"database/sql"
	"math/big"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AccountResponse struct {
	Success bool    `json:"success"`
	Msg     string  `json:"msg"`
	Data    Account `json:"data"`
}

type Account struct {
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	AccountType  AccountType `json:"account_type"`
	Amount       string      `json:"amount"`
	Deleted      bool        `json:"deleted"`
	modelAccount *model.Account
	logger       *logwrap.LogWrap
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

func NewAccount(logger *logwrap.LogWrap, a *model.Account) Account {
	account := Account{
		ID:           a.ID,
		Name:         a.Name,
		modelAccount: a,
		logger:       logger,
	}

	if a.R != nil && a.R.AccountType != nil {
		account.AccountType = AccountType{
			ID:   a.R.AccountType.ID,
			Name: a.R.AccountType.Name,
		}
	}

	return account
}

func (a *Account) SetAmount(ctx context.Context, db *sql.DB) error {
	splits, err := model.Splits(qm.Where("account_id=?", a.modelAccount.ID)).All(ctx, db)
	if err != nil {
		return err
	}

	amount := float64(0)
	for _, split := range splits {
		r := big.NewRat(split.ValueNum, split.ValueDenom)
		f, _ := r.Float64()
		amount = amount + f
	}
	a.Amount = decimal.NewFromFloat(amount).StringFixed(2)
	return nil
}

func (a *Account) SetAttributes(ctx context.Context, db *sql.DB) error {
	if err := a.setDeleted(ctx, db); err != nil {
		return err
	}
	if err := a.setDescription(ctx, db); err != nil {
		return err
	}
	return nil
}

func (a *Account) setDeleted(ctx context.Context, db *sql.DB) error {
	accountAttribute, err := model.AccountAttributes(qm.Where("account_id=? AND name=?", a.modelAccount.ID, "deleted")).One(ctx, db)
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

func (a *Account) setDescription(ctx context.Context, db *sql.DB) error {
	accountAttribute, err := model.AccountAttributes(qm.Where("account_id=? AND name=?", a.modelAccount.ID, "description")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		a.Description = accountAttribute.Value
	}
	return nil
}
