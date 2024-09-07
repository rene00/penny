package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"strings"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	ErrAccountNotFound          = errors.New("account not found")
	ErrAccountAttributeNotFound = errors.New("account attribute not found")
)

type accountService struct {
	ctx              context.Context
	db               *sql.DB
	logger           *logwrap.LogWrap
	defaultResources defaultresources.DefaultResources
}

var account *accountService

func Account() *accountService {
	account = &accountService{}
	return account
}

func (a *accountService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap, defaultResources defaultresources.DefaultResources) {
	a.ctx = ctx
	a.db = db
	a.logger = logger
	a.defaultResources = defaultResources
}

func (a *accountService) getChildAccountType(name string) (model.AccountType, error) {
	var accountType model.AccountType
	var err error

	childAccountTypes, err := a.getAllChildAccountTypes()
	if err != nil {
		return accountType, err
	}

	for _, i := range childAccountTypes {
		if strings.ToLower(i.Name) == strings.ToLower(name) {
			accountType = i
			break
		}
	}

	if (model.AccountType{}) == accountType {
		return accountType, fmt.Errorf("failed to find child account type: %s", name)
	}

	return accountType, nil
}

func (a *accountService) getAllChildAccountTypes() ([]model.AccountType, error) {
	var childAccountTypes []model.AccountType
	var err error
	var q []qm.QueryMod

	q = []qm.QueryMod{
		qm.Where("name=?", "AccountType"),
		qm.Where("parent_id=0"),
	}
	grandParentAccountType, err := model.AccountTypes(q...).One(a.ctx, a.db)
	if err != nil {
		return childAccountTypes, err
	}

	q = []qm.QueryMod{
		qm.Where("parent_id=?", grandParentAccountType.ID),
	}
	parentAccountTypes, err := model.AccountTypes(q...).All(a.ctx, a.db)
	if err != nil {
		return childAccountTypes, err
	}

	for _, parentAccountType := range parentAccountTypes {
		q = []qm.QueryMod{
			qm.Where("parent_id=?", parentAccountType.ID),
		}
		c, err := model.AccountTypes(q...).All(a.ctx, a.db)
		if err != nil {
			return childAccountTypes, err
		}

		for _, childAccountType := range c {
			childAccountTypes = append(childAccountTypes, *childAccountType)
		}
	}

	return childAccountTypes, nil
}

func (a *accountService) GetChildAccountTypes() types.JSResp {
	var resp types.JSResp

	childAccountTypes, err := a.getAllChildAccountTypes()
	if err != nil {
		resp.Msg = "Failed to get child account types"
		a.logger.Error(resp.Msg)
		return resp
	}

	data := []types.AccountType{}
	for _, childAccountType := range childAccountTypes {
		accountType := types.AccountType{
			ID:   childAccountType.ID,
			Name: childAccountType.Name,
		}
		data = append(data, accountType)
	}

	resp.Success = true
	resp.Data = data

	return resp
}

func (a *accountService) UpdateAccount(accountID int64, name, description string, accountTypeID int64) types.JSResp {
	var resp types.JSResp

	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	account, err := model.FindAccount(a.ctx, tx, accountID)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	account.Name = name
	account.AccountTypeID = accountTypeID

	_, err = account.Update(a.ctx, tx, boil.Whitelist("name", "account_type_id"))
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	accountType, err := model.FindAccountType(a.ctx, tx, accountTypeID)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	account.AccountTypeID = accountType.ID

	accountAttribute := model.AccountAttribute{
		ID:        0,
		AccountID: account.ID,
		Name:      "description",
		Value:     description,
	}
	if err := accountAttribute.Upsert(a.ctx, tx, true, []string{"account_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	if err := tx.Commit(); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Account updated"

	return resp
}

func (a *accountService) createAccount(name, description string, accountTypeID int64) (model.Account, error) {
	var err error
	var account model.Account

	accountType, err := model.AccountTypes(qm.Where("id=?", accountTypeID)).One(a.ctx, a.db)
	if err != nil {
		return account, err
	}

	account = model.Account{
		Name:          name,
		EntityID:      a.defaultResources.EntityID(),
		AccountTypeID: accountType.ID,
		ParentID:      a.defaultResources.GrandParentAccountID(),
	}

	if err := account.Upsert(a.ctx, a.db, true, []string{"name", "parent_id", "entity_id", "account_type_id"}, boil.Whitelist("name"), boil.Infer()); err != nil {
		return account, err
	}

	if len(description) != 0 {
		accountAttribute := model.AccountAttribute{
			AccountID: account.ID,
			Name:      "description",
			Value:     description,
		}
		if err := accountAttribute.Insert(a.ctx, a.db, boil.Infer()); err != nil {
			return account, err
		}
	}

	return account, nil
}

func (a *accountService) deleteAccountAttribute(accountID int64, name string) (int64, error) {
	return model.AccountAttributes(qm.Where("account_id=? AND name=?", accountID, name)).DeleteAll(a.ctx, a.db)
}

func (a *accountService) upsertAccountAttribute(accountID int64, name, value string) (model.AccountAttribute, error) {
	accountAttribute := model.AccountAttribute{
		ID:        0,
		AccountID: accountID,
		Name:      name,
		Value:     value,
	}
	if err := accountAttribute.Upsert(a.ctx, a.db, true, []string{"account_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		return accountAttribute, err
	}
	return accountAttribute, nil
}

func (a *accountService) UndeleteAccount(accountID int64) types.AccountResponse {
	resp := types.NewAccountResponse()
	account, err := model.FindAccount(a.ctx, a.db, accountID)
	if err != nil {
		resp.Msg = fmt.Sprintf("Unable to find account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	if _, err := a.deleteAccountAttribute(account.ID, "deleted"); err != nil {
		resp.Msg = fmt.Sprintf("Failed to undelete account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Account undeleted"
	resp.Success = true

	return resp
}

func (a *accountService) DeleteAccount(accountID int64) types.AccountResponse {
	resp := types.NewAccountResponse()
	account, err := model.FindAccount(a.ctx, a.db, accountID)
	if err != nil {
		resp.Msg = fmt.Sprintf("Unable to find account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	splitsExist, err := model.Splits(qm.Where("account_id=?", accountID)).Exists(a.ctx, a.db)
	switch {
	case err != nil:
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	case splitsExist:
		resp.Msg = "Account still assigned to transactions"
		return resp
	}

	if _, err := a.upsertAccountAttribute(account.ID, "deleted", "true"); err != nil {
		resp.Msg = fmt.Sprintf("Failed to delete account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Account deleted"
	resp.Success = true

	return resp
}

func (a *accountService) CreateAccount(name, description string, accountTypeID int64) types.AccountResponse {
	resp := types.NewAccountResponse()

	if len(name) == 0 {
		resp.Msg = "Name must be not be empty"
		return resp
	}

	if accountTypeID == 0 {
		resp.Msg = "Account Type must not be empty"
		return resp
	}

	account, err := a.createAccount(name, description, accountTypeID)
	switch {
	case dberror.IsUniqueConstraint(err):
		resp.Msg = "Account already exists"
		return resp
	case err != nil:
		resp.Msg = fmt.Sprintf("Failed to create Account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	// Check if the account was previously deleted and if it was, remove the account attribute delete.
	accountAttribute, err := model.AccountAttributes(qm.Where("account_id=? AND name=?", account.ID, "deleted")).One(a.ctx, a.db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	default:
		if _, err := accountAttribute.Delete(a.ctx, a.db); err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}
	}

	resp.Data, err = types.NewAccount(a.ctx, a.db, &account)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Account created"
	resp.Success = true
	return resp
}

func (a *accountService) getAccount(id int64) (*model.Account, error) {
	q := []qm.QueryMod{
		qm.Where("account.id=?", id),
		qm.Where("account.entity_id=?", a.defaultResources.EntityID()),
		qm.InnerJoin("account_type on account_type.id = account.account_type_id"),
		qm.Load("AccountType"),
	}
	return model.Accounts(q...).One(a.ctx, a.db)
}

func (a *accountService) GetAccountByName(name string) types.AccountResponse {
	var resp types.AccountResponse
	q := []qm.QueryMod{
		qm.Where("account.name=?", name),
		qm.InnerJoin("account_type on account_type.id = account.account_type_id"),
		qm.Load("AccountType"),
	}
	account, err := model.Accounts(q...).One(a.ctx, a.db)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	resp.Data, err = types.NewAccount(a.ctx, a.db, account)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	resp.Success = true
	return resp
}

func (a *accountService) GetAccount(id int64) types.AccountResponse {
	var err error

	resp := types.NewAccountResponse()

	account, err := a.getAccount(id)
	switch {
	case dberror.IsNoRowsFound(err):
		resp.Msg = "Account not found"
		a.logger.Error(resp.Msg)
		return resp
	case err != nil:
		resp.Msg = fmt.Sprintf("Failed to find account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Data, err = types.NewAccount(a.ctx, a.db, account)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true

	return resp
}

func (a *accountService) GetAccounts() types.JSResp {
	var err error
	var resp types.JSResp
	var q []qm.QueryMod

	q = []qm.QueryMod{
		qm.Where("account.entity_id=?", a.defaultResources.EntityID()),
		qm.Where("account.parent_id=?", a.defaultResources.GrandParentAccountID()),
		qm.InnerJoin("account_type on account_type.id = account.account_type_id"),
		qm.Load("AccountType"),
	}

	accounts, err := model.Accounts(q...).All(a.ctx, a.db)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	data := []types.Account{}
	for _, account := range accounts {
		d, err := types.NewAccount(a.ctx, a.db, account)
		if err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}

		if d.Deleted == true {
			continue
		}

		data = append(data, d)
	}

	resp.Success = true
	resp.Data = data

	return resp
}
