package types

import (
	"context"
	"database/sql"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"

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
	return account
}

func (a *Account) SetDescription(ctx context.Context, db *sql.DB) error {
	accountAttribute, err := model.AccountAttributes(qm.Where("account_id=? AND name=?", a.modelAccount.ID, "description")).One(ctx, db)
	if err != nil {
		return err
	}
	a.Description = accountAttribute.Value
	return nil
}
