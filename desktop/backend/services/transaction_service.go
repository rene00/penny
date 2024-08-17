package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"regexp"
	"strings"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const DefaultCurrency string = "Australian Dollar"

type transactionService struct {
	ctx                    context.Context
	db                     *sql.DB
	logger                 *logwrap.LogWrap
	defaultEntity          *model.Entity
	accountSvc             *accountService
	transactionImporterSvc *transactionImporterService
}

var transaction *transactionService

func Transaction() *transactionService {
	transaction = &transactionService{}
	return transaction
}

func (t *transactionService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap, transactionImporterSvc *transactionImporterService) {
	t.ctx = ctx
	t.db = db
	t.logger = logger
	t.transactionImporterSvc = transactionImporterSvc

	entity, err := model.Entities(qm.Where("name=?", "Default")).One(t.ctx, t.db)
	if err != nil {
		t.logger.Error(fmt.Sprintf("Failed to find default entity: %s", err.Error()))
	}
	t.defaultEntity = entity
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

	if err := data.SetAttributes(t.ctx, t.db); err != nil {
		resp.Msg = fmt.Sprintf("Failed to set attributes for transaction %d: %s", transaction.ID, err.Error())
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
			resp.Msg = fmt.Sprintf("Failed to get splits for transaction %d: %s", i.ID, err.Error())
			t.logger.Error(resp.Msg)
			return resp
		}

		if err := transaction.SetAttributes(t.ctx, t.db); err != nil {
			resp.Msg = fmt.Sprintf("Failed to set attributes for transaction %d: %s", i.ID, err.Error())
			t.logger.Error(resp.Msg)
			return resp
		}

		if transaction.Deleted == true {
			continue
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

func (t *transactionService) deleteTransaction(transaction *model.Transaction) error {
	transactionAttribute := model.TransactionsAttribute{
		TransactionsID: transaction.ID,
		Name:           "deleted",
		Value:          "true",
	}
	return transactionAttribute.Upsert(t.ctx, t.db, true, []string{"transactions_id", "name"}, boil.Whitelist("name"), boil.Infer())
}

func (t *transactionService) undeleteTransaction(transaction *model.Transaction) error {
	_, err := model.TransactionsAttributes(qm.Where("transactions_id=? AND name=?", transaction.ID, "deleted")).DeleteAll(t.ctx, t.db)
	if err != nil {
		return err
	}
	return nil
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

func (t *transactionService) UndeleteTransaction(id int64) types.TransactionResponse {
	var resp types.TransactionResponse

	transaction, err := t.getTransaction(id)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to find transaction: %s", err)
		return resp
	}

	if err := t.undeleteTransaction(transaction); err != nil {
		resp.Msg = fmt.Sprintf("Failed to undelete transaction: %s", err)
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Transaction undeleted"
	resp.Success = true

	return resp
}

func (t *transactionService) DeleteTransaction(id int64) types.TransactionResponse {
	var resp types.TransactionResponse

	transaction, err := t.getTransaction(id)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to find transaction: %s", err)
		return resp
	}

	if err := t.deleteTransaction(transaction); err != nil {
		resp.Msg = fmt.Sprintf("Failed to delete transaction: %s", err)
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Transaction deleted"
	resp.Success = true

	return resp
}

// ImportTransactions accepts string and returns JSResp. String is a base64
// string from JS which is reader from JS FileReader() API. String from JS will
// start with `data:application/octet-stream;base64,`
func (t *transactionService) ImportTransactions(s string) types.JSResp {
	var resp types.JSResp

	sArray := strings.Split(s, ",")
	if len(sArray) != 2 {
		resp.Msg = "Failed to split transaction string"
		t.logger.Error(resp.Msg)
		return resp
	}

	decoded, err := base64.StdEncoding.DecodeString(sArray[1])
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to decode: %s", err.Error())
		t.logger.Error(resp.Msg)
		return resp
	}

	h := sha256.New()
	h.Write(decoded)

	req := TransactionImporterRequest{
		Path:     ":blob:",
		Contents: decoded,
		Checksum: h.Sum(nil),
	}
	t.transactionImporterSvc.TransactionImporterRequestC <- req

	resp.Msg = "Imported transactions"
	resp.Success = true
	return resp
}
