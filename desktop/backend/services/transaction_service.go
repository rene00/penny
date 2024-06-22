package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"pennyapp/backend/internal/hashers"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"regexp"
	"strings"
	"time"

	"github.com/aclindsa/ofxgo"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const DefaultCurrency string = "Australian Dollar"

type transactionService struct {
	ctx           context.Context
	db            *sql.DB
	logger        *logwrap.LogWrap
	defaultEntity *model.Entity
	accountSvc    *accountService
}

var transaction *transactionService

func Transaction() *transactionService {
	transaction = &transactionService{}
	return transaction
}

func (t *transactionService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	t.ctx = ctx
	t.db = db
	t.logger = logger

	entity, err := model.Entities(qm.Where("name=?", "Default")).One(t.ctx, t.db)
	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to find default entity: %s", err.Error()))
	}
	t.defaultEntity = entity

	return
}

// TODO: refactor to AddHandler() with Opts and support WithAccountService() opt func.
func (t *transactionService) AddAccountSvcHandler(a *accountService) {
	t.accountSvc = a
}

func (t *transactionService) getTransaction(id int64) (*model.Transaction, error) {
	q := []qm.QueryMod{
		qm.Where("transactions.id=?", id),
		qm.Where("transactions.entity_id=?", t.defaultEntity.ID),
	}
	return model.Transactions(q...).One(t.ctx, t.db)
}

func (t *transactionService) GetTransaction(id int64) types.TransactionResponse {

	resp := types.NewTransactionResponse()

	transaction, err := t.getTransaction(id)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	data, err := types.NewTransaction(t.logger, transaction)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	if err := data.SetSplits(t.ctx, t.db); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Data = data

	return resp

}

func (t *transactionService) GetTransactions() types.JSResp {
	var err error
	var resp types.JSResp
	var q []qm.QueryMod

	q = []qm.QueryMod{
		qm.Where("transactions.entity_id=?", t.defaultEntity.ID),
		qm.OrderBy("transactions.date DESC"),
	}
	transactions, err := model.Transactions(q...).All(t.ctx, t.db)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	data := []types.Transaction{}

	for _, i := range transactions {
		transaction, err := types.NewTransaction(t.logger, i)
		if err != nil {
			resp.Msg = err.Error()
			t.logger.Error(resp.Msg)
			return resp
		}

		if err := transaction.SetSplits(t.ctx, t.db); err != nil {
			t.logger.Error(fmt.Sprintf("failed to set splits for transaction %d: %s", transaction.ID, err.Error()))
		}

		transaction.SetAmount()
		transaction.SetDisplay()

		data = append(data, transaction)
	}

	resp.Success = true
	resp.Data = data

	return resp
}

func (t *transactionService) UpdateTransaction(transactionID int64, memo, date string) types.JSResp {
	var resp types.JSResp
	var err error

	if memo == "" {
		resp.Msg = "Memo cant be empty"
		t.logger.Error(resp.Msg)
		return resp
	}

	if date == "" {
		resp.Msg = "Date cant be empty"
		t.logger.Error(resp.Msg)
		return resp
	}

	tx, err := t.db.BeginTx(t.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	transaction, err := model.FindTransaction(t.ctx, tx, transactionID)
	if err != nil {
		resp.Msg = "Unable to find Transaction"
		t.logger.Error(fmt.Sprintf("%s: %s", err.Error(), resp.Msg))
		return resp
	}
	transaction.Memo = memo

	d, err := GetSQLiteDateFromVueJSDate(date)
	if err != nil {
		resp.Msg = "Date format must be YYYY-MM-DD"
		t.logger.Error(fmt.Sprintf("%s: %s", err.Error(), resp.Msg))
		return resp
	}
	transaction.Date = d

	_, err = transaction.Update(t.ctx, tx, boil.Whitelist("memo", "date"))
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	if err := tx.Commit(); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Transaction updated"

	return resp
}

func (t *transactionService) getDefaultCurrency() (*model.Currency, error) {
	return model.Currencies(qm.Where("name=?", DefaultCurrency)).One(t.ctx, t.db)
}

func (t *transactionService) createTransaction(memo, date string) (model.Transaction, error) {
	var transaction model.Transaction
	var err error

	currency, err := t.getDefaultCurrency()
	if err != nil {
		return transaction, err
	}

	// vuejs datepicker sends in date string like this
	// 2024-05-22T20:36:00.000Z. I wasn't sure how to remove time from date
	// string. Instead, truncate time from date string and pass into time.Parse
	// as YYYY-MM-DD.
	re := regexp.MustCompile(`(?P<date>^\d{4}-\d{2}-\d{2})T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)
	tpl := "${date}"
	matches := re.FindStringSubmatchIndex(date)
	result := []byte{}
	result = re.ExpandString(result, tpl, date, matches)

	d, err := time.Parse("2006-01-02", string(result))
	if err != nil {
		return transaction, err
	}

	transaction = model.Transaction{
		Memo:       memo,
		Date:       d.Format("2006-01-02 03:04:05.000"),
		CurrencyID: currency.ID,
		EntityID:   t.defaultEntity.ID,
	}

	if err := transaction.Insert(t.ctx, t.db, boil.Infer()); err != nil {
		return transaction, err
	}

	return transaction, nil
}

func (t *transactionService) CreateTransaction(memo, date string) types.TransactionResponse {
	var resp types.TransactionResponse

	if len(memo) == 0 {
		resp.Msg = "Memo must not be empty"
		return resp
	}

	transaction, err := t.createTransaction(memo, date)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to create transaction: %s", err)
		t.logger.Error(resp.Msg)
		return resp
	}

	currency, err := t.getDefaultCurrency()
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to get default currency: %s", err)
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Created transaction"
	resp.Success = true
	resp.Data = types.Transaction{
		ID:   transaction.ID,
		Memo: transaction.Memo,
		Date: transaction.Date,
		Currency: types.Currency{
			ID:   currency.ID,
			Name: currency.Name,
		},
	}

	return resp
}

// ImportTransactions accepts string and returns JSResp. String is a base64
// string from JS which is reader from JS FileReader() API. String from JS will
// start with `data:application/octet-stream;base64,`
func (t *transactionService) ImportTransactions(s string) types.JSResp {
	var resp types.JSResp

	sArray := strings.Split(s, ",")
	if len(sArray) != 2 {
		resp.Msg = fmt.Sprintf("Failed to split transaction string")
		t.logger.Error(resp.Msg)
		return resp
	}

	decoded, err := base64.StdEncoding.DecodeString(sArray[1])
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to decode: %s", err.Error())
		t.logger.Error(resp.Msg)
		return resp
	}

	ofxResp, err := ofxgo.ParseResponse(strings.NewReader(string(decoded)))
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to parse: %s", err.Error())
		t.logger.Error(resp.Msg)
		return resp
	}

	switch {
	case len(ofxResp.Bank) > 0:
		for _, msg := range ofxResp.Bank {
			stmt, ok := msg.(*ofxgo.StatementResponse)
			if !ok {
				resp.Msg = fmt.Sprintf("failed to parse bank account response: %s", err.Error())
				t.logger.Error(resp.Msg)
				return resp
			}

			account, err := t.accountSvc.getBankAccount(stmt.BankAcctFrom)
			switch {
			case err == ErrAccountNotFound:
				account, err = t.accountSvc.CreateBankAccount(stmt.BankAcctFrom)
				if err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}
			case err != nil:
				resp.Msg = err.Error()
				t.logger.Error(resp.Msg)
				return resp
			default:
			}

			currency, err := t.getCurrencyFromOfxCurrencySymbol(stmt.CurDef)
			if err != nil {
				resp.Msg = err.Error()
				return resp
			}

			for _, i := range stmt.BankTranList.Transactions {
				transaction := model.Transaction{
					EntityID:   t.defaultEntity.ID,
					Memo:       i.Memo.String(),
					CurrencyID: currency.ID,
					Date:       i.DtPosted.Format("2006-01-02 03:04:05.000"),
				}

				v1Hash := hashers.NewTransactionHashV1()
				transactionHash := v1Hash.Hash(transaction)
				hashType, err := model.TransactionsHashTypes(qm.Where("name=?", v1Hash.String())).One(t.ctx, t.db)
				if err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}

				q := []qm.QueryMod{
					qm.Where("hash=?", fmt.Sprintf("%x", transactionHash.Sum(nil))),
					qm.Where("transactions_hash_type_id=?", hashType.ID),
				}
				exists, err := model.TransactionsHashes(q...).Exists(t.ctx, t.db)
				if err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}
				if exists {
					continue
				}

				if err := transaction.Insert(t.ctx, t.db, boil.Infer()); err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}

				th := model.TransactionsHash{
					TransactionsID:         transaction.ID,
					TransactionsHashTypeID: hashType.ID,
					Hash:                   fmt.Sprintf("%x", transactionHash.Sum(nil)),
				}
				if err := th.Insert(t.ctx, t.db, boil.Infer()); err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}

				var attr model.TransactionsAttribute
				transactionAttributes := []model.TransactionsAttribute{}

				if len(i.FiTID.String()) != 0 {
					attr = model.TransactionsAttribute{
						TransactionsID: transaction.ID,
						Name:           "fitid",
						Value:          i.FiTID.String(),
					}
					transactionAttributes = append(transactionAttributes, attr)
				}

				if len(i.TrnType.String()) != 0 {
					attr = model.TransactionsAttribute{
						TransactionsID: transaction.ID,
						Name:           "trntype",
						Value:          i.TrnType.String(),
					}
					transactionAttributes = append(transactionAttributes, attr)
				}

				for _, ii := range transactionAttributes {
					if err := ii.Insert(t.ctx, t.db, boil.Infer()); err != nil {
						resp.Msg = err.Error()
						t.logger.Error(resp.Msg)
						return resp
					}
				}

				splits := []model.Split{
					{
						TransactionsID: transaction.ID,
						AccountID:      account.ID,
						ValueNum:       i.TrnAmt.Num().Int64(),
						ValueDenom:     i.TrnAmt.Denom().Int64(),
					},
				}
				for _, ii := range splits {
					if err := ii.Insert(t.ctx, t.db, boil.Infer()); err != nil {
						resp.Msg = err.Error()
						t.logger.Error(resp.Msg)
						return resp
					}
				}
			}
		}
	case len(ofxResp.CreditCard) > 0:
		for _, msg := range ofxResp.CreditCard {
			stmt, ok := msg.(*ofxgo.CCStatementResponse)
			if !ok {
				resp.Msg = fmt.Sprintf("failed to parse credit card response: %s", err.Error())
				t.logger.Error(resp.Msg)
			}

			account, err := t.accountSvc.getCreditCardAccount(stmt.CCAcctFrom)
			if err != nil {
				resp.Msg = err.Error()
				t.logger.Error(resp.Msg)
				return resp
			}

			currency, err := t.getCurrencyFromOfxCurrencySymbol(stmt.CurDef)
			if err != nil {
				resp.Msg = err.Error()
				return resp
			}

			for _, i := range stmt.BankTranList.Transactions {
				transaction := model.Transaction{
					EntityID:   t.defaultEntity.ID,
					Memo:       i.Memo.String(),
					CurrencyID: currency.ID,
					Date:       i.DtPosted.Format("2006-01-02 03:04:05.000"),
				}

				v1Hash := hashers.NewTransactionHashV1()
				transactionHash := v1Hash.Hash(transaction)
				hashType, err := model.TransactionsHashTypes(qm.Where("name=?", v1Hash.String())).One(t.ctx, t.db)
				if err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}

				q := []qm.QueryMod{
					qm.Where("hash=?", fmt.Sprintf("%x", transactionHash.Sum(nil))),
					qm.Where("transactions_hash_type_id=?", hashType.ID),
				}
				exists, err := model.TransactionsHashes(q...).Exists(t.ctx, t.db)
				if err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}
				if exists {
					continue
				}

				if err := transaction.Insert(t.ctx, t.db, boil.Infer()); err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}

				th := model.TransactionsHash{
					TransactionsID:         transaction.ID,
					TransactionsHashTypeID: hashType.ID,
					Hash:                   fmt.Sprintf("%x", transactionHash.Sum(nil)),
				}
				if err := th.Insert(t.ctx, t.db, boil.Infer()); err != nil {
					resp.Msg = err.Error()
					t.logger.Error(resp.Msg)
					return resp
				}

				var attr model.TransactionsAttribute
				transactionAttributes := []model.TransactionsAttribute{}

				if len(i.FiTID.String()) != 0 {
					attr = model.TransactionsAttribute{
						TransactionsID: transaction.ID,
						Name:           "fitid",
						Value:          i.FiTID.String(),
					}
					transactionAttributes = append(transactionAttributes, attr)
				}

				if len(i.TrnType.String()) != 0 {
					attr = model.TransactionsAttribute{
						TransactionsID: transaction.ID,
						Name:           "trntype",
						Value:          i.TrnType.String(),
					}
					transactionAttributes = append(transactionAttributes, attr)
				}

				for _, ii := range transactionAttributes {
					if err := ii.Insert(t.ctx, t.db, boil.Infer()); err != nil {
						resp.Msg = err.Error()
						t.logger.Error(resp.Msg)
						return resp
					}
				}

				splits := []model.Split{
					{
						TransactionsID: transaction.ID,
						AccountID:      account.ID,
						ValueNum:       i.TrnAmt.Num().Int64(),
						ValueDenom:     i.TrnAmt.Denom().Int64(),
					},
				}
				for _, ii := range splits {
					if err := ii.Insert(t.ctx, t.db, boil.Infer()); err != nil {
						resp.Msg = err.Error()
						t.logger.Error(resp.Msg)
						return resp
					}
				}
			}

		}
	default:
		resp.Msg = fmt.Sprintf("OFX file not supported")
		return resp
	}

	resp.Msg = "Imported transactions"
	resp.Success = true
	return resp
}

func (t transactionService) getCurrencyFromOfxCurrencySymbol(curDef ofxgo.CurrSymbol) (*model.Currency, error) {
	var currency *model.Currency
	var err error

	currencyAttribute, err := model.CurrencyAttributes(
		[]qm.QueryMod{
			qm.Where("name=?", "code"),
			qm.Where("value=?", fmt.Sprintf("%s", curDef)),
		}...,
	).One(t.ctx, t.db)
	if err != nil {
		t.logger.Error(err.Error())
		return currency, err
	}

	currency, err = model.FindCurrency(t.ctx, t.db, currencyAttribute.ID)
	if err != nil {
		t.logger.Error(err.Error())
		return currency, err
	}

	return currency, nil
}
