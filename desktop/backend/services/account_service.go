package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"pennyapp/backend/internal/hashers"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"regexp"
	"strings"

	"github.com/aclindsa/ofxgo"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	ErrAccountNotFound          = errors.New("account not found")
	ErrAccountAttributeNotFound = errors.New("account attribute not found")
)

type accountService struct {
	ctx                context.Context
	db                 *sql.DB
	logger             *logwrap.LogWrap
	defaultEntity      *model.Entity
	grandParentAccount *model.Account
}

var account *accountService

func Account() *accountService {
	account = &accountService{}
	return account
}

func (a *accountService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	a.ctx = ctx
	a.db = db
	a.logger = logger

	entity, err := model.Entities(qm.Where("name=?", "Default")).One(a.ctx, a.db)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to find default entity: %s", err.Error()))
	}
	a.defaultEntity = entity

	q := []qm.QueryMod{
		qm.Where("name=?", "Account"),
		qm.Where("parent_id=0"),
	}
	account, err := model.Accounts(q...).One(a.ctx, a.db)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to find grand parent account: %s", err.Error()))
	}
	a.grandParentAccount = account
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
		EntityID:      a.defaultEntity.ID,
		AccountTypeID: accountType.ID,
		ParentID:      a.grandParentAccount.ID,
	}
	if err := account.Insert(a.ctx, a.db, boil.Infer()); err != nil {
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

func (a *accountService) CreateAccount(name, description string, accountTypeID int64) (resp types.JSResp) {
	if len(name) == 0 {
		resp.Msg = "Name must be not be empty"
		return resp
	}

	if accountTypeID == 0 {
		resp.Msg = "Account Type must not be empty"
		return resp
	}

	if _, err := a.createAccount(name, description, accountTypeID); err != nil {
		resp.Msg = fmt.Sprintf("Failed to create account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Account created"
	resp.Success = true
	return resp
}

func (a *accountService) GetAccount(id int64) types.JSResp {
	var err error
	var resp types.JSResp
	var q []qm.QueryMod

	q = []qm.QueryMod{
		qm.Where("account.id=?", id),
		qm.Where("account.entity_id=?", a.defaultEntity.ID),
		qm.InnerJoin("account_type on account_type.id = account.account_type_id"),
		qm.Load("AccountType"),
	}
	account, err := model.Accounts(q...).One(a.ctx, a.db)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to find account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	accountResp := types.Account{
		ID:   account.ID,
		Name: account.Name,
		AccountType: types.AccountType{
			ID:   account.R.AccountType.ID,
			Name: account.R.AccountType.Name,
		},
	}

	accountAttributes, err := model.AccountAttributes(qm.Where("account_id=?", account.ID)).All(a.ctx, a.db)
	for _, accountAttribute := range accountAttributes {
		switch strings.ToLower(accountAttribute.Name) {
		case "description":
			accountResp.Description = accountAttribute.Value
		default:
		}
	}

	splits, err := model.Splits(qm.Where("account_id=?", account.ID)).All(a.ctx, a.db)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	amount := float64(0)
	for _, split := range splits {
		r := big.NewRat(split.ValueNum, split.ValueDenom)
		f, _ := r.Float64()
		amount = amount + f
	}
	accountResp.Amount = decimal.NewFromFloat(amount).StringFixed(2)

	resp.Success = true
	resp.Data = accountResp

	return resp
}

func (a *accountService) GetAccounts() types.JSResp {
	var err error
	var resp types.JSResp
	var q []qm.QueryMod

	q = []qm.QueryMod{
		qm.Where("name=?", "Account"),
		qm.Where("entity_id=?", a.defaultEntity.ID),
		qm.Where("parent_id=0"),
	}
	grandParentAccount, err := model.Accounts(q...).One(a.ctx, a.db)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to find grand parent account: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	q = []qm.QueryMod{
		qm.Where("account.entity_id=?", a.defaultEntity.ID),
		qm.Where("account.parent_id=?", grandParentAccount.ID),
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
	for _, i := range accounts {
		account := types.Account{
			ID:   i.ID,
			Name: i.Name,
			AccountType: types.AccountType{
				ID:   i.R.AccountType.ID,
				Name: i.R.AccountType.Name,
			},
		}

		accountAttributes, err := model.AccountAttributes(qm.Where("account_id = ?", i.ID)).All(a.ctx, a.db)
		if err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}
		for _, accountAttribute := range accountAttributes {
			switch strings.ToLower(accountAttribute.Name) {
			case "description":
				account.Description = accountAttribute.Value
			default:
			}
		}

		splits, err := model.Splits(qm.Where("account_id=?", i.ID)).All(a.ctx, a.db)
		if err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}

		amount := float64(0)
		for _, split := range splits {
			r := big.NewRat(split.ValueNum, split.ValueDenom)
			f, _ := r.Float64()
			amount = amount + f
		}
		account.Amount = decimal.NewFromFloat(amount).StringFixed(2)

		data = append(data, account)

	}

	resp.Success = true
	resp.Data = data

	return resp
}

func (a *accountService) CreateBankAccount(bankAccount ofxgo.BankAcct) (model.Account, error) {
	var account model.Account
	var err error

	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	accountType, err := model.AccountTypes(qm.Where("name=?", "Asset")).One(a.ctx, tx)
	if err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	account = model.Account{
		Name:          fmt.Sprintf("Bank Account %s%s", bankAccount.BankID.String(), bankAccount.AcctID.String()),
		AccountTypeID: accountType.ID,
		EntityID:      a.defaultEntity.ID,
		ParentID:      a.grandParentAccount.ID,
	}

	if err := account.Insert(a.ctx, tx, boil.Infer()); err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	accountAttributes := []model.AccountAttribute{
		{
			AccountID: account.ID,
			Name:      "bsb",
			Value:     bankAccount.BankID.String(),
		},
		{
			AccountID: account.ID,
			Name:      "account_number",
			Value:     bankAccount.AcctID.String(),
		},
		{
			AccountID: account.ID,
			Name:      "ofx_accttype",
			Value:     bankAccount.AcctType.String(),
		},
	}

	for _, i := range accountAttributes {
		if err := i.Insert(a.ctx, tx, boil.Infer()); err != nil {
			a.logger.Error(err.Error())
			return account, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		a.logger.Error(err.Error())
		return account, err
	}

	return account, nil
}

func (a accountService) getSurpressedCreditCardNumber(s string) (string, error) {
	re := regexp.MustCompile(`^\d{16}$`)
	if !re.MatchString(s) {
		return "", fmt.Errorf("not a credit card")
	}

	rs := []rune(s)
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = 'X'
	}

	return string(rs), nil
}

func (a accountService) CreateCreditCardAccount(cc ofxgo.CCAcct) (model.Account, error) {
	var account model.Account
	var err error

	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	accountType, err := model.AccountTypes(qm.Where("name=?", "Liability")).One(a.ctx, tx)
	if err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	ccSurpressed, err := a.getSurpressedCreditCardNumber(cc.AcctID.String())
	if err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	account = model.Account{
		Name:          fmt.Sprintf("Credit Card %s", ccSurpressed),
		AccountTypeID: accountType.ID,
		EntityID:      a.defaultEntity.ID,
		ParentID:      a.grandParentAccount.ID,
	}

	if err := account.Insert(a.ctx, tx, boil.Infer()); err != nil {
		a.logger.Error(err.Error())
		return account, err
	}

	accountAttributes := []model.AccountAttribute{
		{
			AccountID: account.ID,
			Name:      "cc_account_id_hash",
			Value:     getCreditCardHash(cc),
		},
	}

	for _, i := range accountAttributes {
		if err := i.Insert(a.ctx, tx, boil.Infer()); err != nil {
			a.logger.Error(err.Error())
			return account, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		a.logger.Error(err.Error())
		return account, err
	}

	return account, nil
}

func (a accountService) getAccountAttributesWithBankAccountBSB(bankAccountBSB string) (model.AccountAttributeSlice, error) {
	q := []qm.QueryMod{
		qm.Where("name=?", "bsb"),
		qm.Where("value=?", bankAccountBSB),
	}
	return model.AccountAttributes(q...).All(a.ctx, a.db)
}

func (a accountService) getAccountAttributeWithCCAccountIDHash(h hash.Hash) (*model.AccountAttribute, error) {
	q := []qm.QueryMod{
		qm.Where("aa.name=?", "cc_account_id_hash"),
		qm.Where("aa.value=?", fmt.Sprintf("%x", h.Sum(nil))),
		qm.InnerJoin("account a on a.id = aa.account_id"),
		qm.From("account_attributes as aa"),
		qm.Load("Account"),
	}
	return model.AccountAttributes(q...).One(a.ctx, a.db)
}

func (a accountService) getCreditCardAccount(cc ofxgo.CCAcct) (model.Account, error) {
	var account model.Account
	var err error

	h := hashers.NewCreditCardHash()
	creditCardHash := h.Hash(cc.AcctID.String())

	accountAttributesWithCCAccountIDHash, err := a.getAccountAttributeWithCCAccountIDHash(creditCardHash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// CC account does not exist so create it.
			account, err = a.CreateCreditCardAccount(cc)
			if err != nil {
				a.logger.Error(err.Error())
				return account, err
			}
			return account, nil
		default:
			a.logger.Error(err.Error())
			return account, err
		}
	}
	account = *accountAttributesWithCCAccountIDHash.R.GetAccount()

	return account, nil
}

func (a accountService) getBankAccount(bankAccount ofxgo.BankAcct) (model.Account, error) {
	var account model.Account
	var err error

	accountAttributesWithBSB, err := a.getAccountAttributesWithBankAccountBSB(bankAccount.BankID.String())
	if err != nil {
		return account, err
	}

	for _, i := range accountAttributesWithBSB {
		// All accounts which share the same account number and BSB.
		q := []qm.QueryMod{
			qm.Where("aa.name=?", "account_number"),
			qm.Where("aa.value=?", bankAccount.AcctID.String()),
			qm.Where("aa.account_id=?", i.AccountID),
			qm.InnerJoin("account a on a.id = aa.account_id"),
			qm.From("account_attributes as aa"),
			qm.Load("Account"),
		}
		accountAttributesWithBSBAndAccountNumber, err := model.AccountAttributes(q...).One(a.ctx, a.db)
		if err != nil {
			return account, err
		}
		switch {
		case accountAttributesWithBSBAndAccountNumber == nil:
			// No accounts found, move to next account.
			continue
		default:
			account = *accountAttributesWithBSBAndAccountNumber.R.GetAccount()
			break
		}
	}

	if (model.Account{}) == account {
		return account, ErrAccountNotFound
	}

	return account, err
}

func getCreditCardHash(cc ofxgo.CCAcct) string {
	return fmt.Sprintf("%x", hashers.NewCreditCardHash().Hash(cc.AcctID.String()).Sum(nil))
}
