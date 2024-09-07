package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"pennyapp/backend/config"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/internal/accounts"
	"pennyapp/backend/internal/hashers"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"time"

	"github.com/aclindsa/ofxgo"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type transactionImporterService struct {
	ctx                         context.Context
	db                          *sql.DB
	conf                        config.Config
	logger                      *logwrap.LogWrap
	defaultResources            defaultresources.DefaultResources
	accountSvc                  *accountService
	TransactionImporterRequestC chan TransactionImporterRequest
}

var transactionImporter *transactionImporterService

func TransactionImporter() *transactionImporterService {
	return &transactionImporterService{}
}

func (t *transactionImporterService) Start(ctx context.Context, conf config.Config, db *sql.DB, logger *logwrap.LogWrap, defaultResources defaultresources.DefaultResources) {
	t.ctx = ctx
	t.conf = conf
	t.logger = logger
	t.db = db
	t.defaultResources = defaultResources
	t.TransactionImporterRequestC = make(chan TransactionImporterRequest)

	go func() { t.importer() }()

	dirWatcherPipeline := func() error {
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := t.dirWatcher(); err != nil {
					t.logger.Error(fmt.Sprintf("failed calling directory watcher: %s", err.Error()))
				}
			case <-t.ctx.Done():
				return t.ctx.Err()
			}
		}
	}

	go func() {
		if err := dirWatcherPipeline(); err != nil {
			t.logger.Error(fmt.Sprintf("failed done cancelation for directory watcher: %s", err.Error()))
		}
	}()
}

func (t *transactionImporterService) dirWatcher() error {
	return filepath.Walk(t.conf.ImportDir, func(path string, info os.FileInfo, walkErr error) error {
		var err error

		if walkErr != nil {
			return walkErr
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		contents, err := os.OpenFile(path, os.O_RDONLY, 0600)
		if err != nil {
			return err
		}

		req, err := NewTransactionImporterRequest(filepath.Join(t.conf.ImportDir, info.Name()), contents)
		if err != nil {
			return err
		}
		t.TransactionImporterRequestC <- req

		return nil
	})
}

func (t *transactionImporterService) importer() {
	for req := range t.TransactionImporterRequestC {
		var err error

		t.logger.Info(fmt.Sprintf("received new transaction importer request: %s", req.Path))

		transactionImport := model.TransactionImporter{
			StatusID: t.defaultResources.TransactionImporterStatusPending().ID,
			Filename: filepath.Base(req.Path),
			Filesize: req.Size,
			Checksum: fmt.Sprintf("%x", req.Checksum),
		}

		// Check if transaction import exists as "completed". If so, do not import.
		exists, err := model.TransactionImporters(qm.Where("checksum=? AND status_id=?", transactionImport.Checksum, t.defaultResources.TransactionImporterStatusCompleted().ID)).Exists(t.ctx, t.db)
		if err != nil {
			t.logger.Error(fmt.Sprintf("failed to check if previous transaction import exists: %s", err.Error()))
			continue
		}
		if exists {
			if _, err := os.Stat(req.Path); err == nil {
				if err := os.Remove(req.Path); err != nil {
					t.logger.Error(fmt.Sprintf("failed removing auto transaction importer file: %s", err.Error()))
				}
			}
			t.logger.Info(fmt.Sprintf("auto transaction importer file already successfully imported: %s", req.Path))
			continue
		}

		if err = transactionImport.Insert(t.ctx, t.db, boil.Infer()); err != nil {
			t.logger.Error(fmt.Sprintf("failed to read previous transaction import: %s", err.Error()))
			continue
		}

		if err := t.Do(req); err != nil {
			t.logger.Error(fmt.Sprintf("failed to process transaction importer request: %s", err.Error()))
		}

		if _, err := os.Stat(req.Path); err == nil {
			if err := os.Remove(req.Path); err != nil {
				t.logger.Error(fmt.Sprintf("failed removing auto transaction importer file: %s", err.Error()))
			}
		}

		transactionImport.StatusID = t.defaultResources.TransactionImporterStatusCompleted().ID
		if _, err := transactionImport.Update(t.ctx, t.db, boil.Infer()); err != nil {
			t.logger.Error(fmt.Sprintf("failed updating transaction importer: %s", err.Error()))
		}
	}
}

func (t *transactionImporterService) ofxParse(r io.Reader) (*ofxgo.Response, error) {
	return ofxgo.ParseResponse(r)
}

func (t *transactionImporterService) getCurrencyFromOfxCurrencySymbol(curDef ofxgo.CurrSymbol) (*model.Currency, error) {
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

func (t transactionImporterService) getAccountAttributesWithBankAccountBSB(bankAccountBSB string) (model.AccountAttributeSlice, error) {
	q := []qm.QueryMod{
		qm.Where("name=?", "bsb"),
		qm.Where("value=?", bankAccountBSB),
	}
	return model.AccountAttributes(q...).All(t.ctx, t.db)
}

func (t transactionImporterService) getCreditCardAccount(cc ofxgo.CCAcct) (model.Account, error) {
	var account model.Account
	var err error

	h := hashers.NewCreditCardHash()
	creditCardHash := h.Hash(cc.AcctID.String())

	accountAttributesWithCCAccountIDHash, err := t.getAccountAttributeWithCCAccountIDHash(creditCardHash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// CC account does not exist so create it.
			account, err = accounts.CreateCreditCardAccount(t.ctx, t.db, t.defaultResources, cc)
			if err != nil {
				t.logger.Error(err.Error())
				return account, err
			}
			return account, nil
		default:
			t.logger.Error(err.Error())
			return account, err
		}
	}
	account = *accountAttributesWithCCAccountIDHash.R.GetAccount()

	return account, nil
}

func (t transactionImporterService) getAccountAttributeWithCCAccountIDHash(h hash.Hash) (*model.AccountAttribute, error) {
	q := []qm.QueryMod{
		qm.Where("aa.name=?", "cc_account_id_hash"),
		qm.Where("aa.value=?", fmt.Sprintf("%x", h.Sum(nil))),
		qm.InnerJoin("account a on a.id = aa.account_id"),
		qm.From("account_attributes as aa"),
		qm.Load("Account"),
	}
	return model.AccountAttributes(q...).One(t.ctx, t.db)
}

func (t transactionImporterService) getBankAccount(bankAccount ofxgo.BankAcct) (model.Account, error) {
	var account model.Account
	var err error

	accountAttributesWithBSB, err := t.getAccountAttributesWithBankAccountBSB(bankAccount.BankID.String())
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
		accountAttributesWithBSBAndAccountNumber, err := model.AccountAttributes(q...).One(t.ctx, t.db)
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

func (t *transactionImporterService) ofxRespBank(msg ofxgo.Message) error {
	stmt, ok := msg.(*ofxgo.StatementResponse)
	if !ok {
		return fmt.Errorf("failed to parse bank account response")
	}

	account, err := t.getBankAccount(stmt.BankAcctFrom)
	switch {
	case err == ErrAccountNotFound:
		account, err = t.createBankAccount(stmt.BankAcctFrom)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	default:
	}

	currency, err := t.getCurrencyFromOfxCurrencySymbol(stmt.CurDef)
	if err != nil {
		return err
	}

	for _, i := range stmt.BankTranList.Transactions {
		transaction := model.Transaction{
			EntityID:   t.defaultResources.EntityID(),
			Memo:       i.Memo.String(),
			CurrencyID: currency.ID,
			Date:       i.DtPosted.Format("2006-01-02 03:04:05.000"),
		}

		v1Hash := hashers.NewTransactionHashV1()
		transactionHash := v1Hash.Hash(transaction)
		hashType, err := model.TransactionsHashTypes(qm.Where("name=?", v1Hash.String())).One(t.ctx, t.db)
		if err != nil {
			return err
		}

		q := []qm.QueryMod{
			qm.Where("hash=?", fmt.Sprintf("%x", transactionHash.Sum(nil))),
			qm.Where("transactions_hash_type_id=?", hashType.ID),
		}
		exists, err := model.TransactionsHashes(q...).Exists(t.ctx, t.db)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if err := transaction.Insert(t.ctx, t.db, boil.Infer()); err != nil {
			return err
		}

		th := model.TransactionsHash{
			TransactionsID:         transaction.ID,
			TransactionsHashTypeID: hashType.ID,
			Hash:                   fmt.Sprintf("%x", transactionHash.Sum(nil)),
		}
		if err := th.Insert(t.ctx, t.db, boil.Infer()); err != nil {
			return err
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
				return err
			}
		}

		imbalancedAccount, err := model.Accounts(qm.Where("name=?", "Imbalanced")).One(t.ctx, t.db)
		if err != nil {
			return err
		}

		splits := []model.Split{
			{
				TransactionsID: transaction.ID,
				AccountID:      account.ID,
				ValueNum:       i.TrnAmt.Num().Int64(),
				ValueDenom:     i.TrnAmt.Denom().Int64(),
			},
			{
				TransactionsID: transaction.ID,
				AccountID:      imbalancedAccount.ID,
				ValueNum:       -i.TrnAmt.Num().Int64(),
				ValueDenom:     -i.TrnAmt.Denom().Int64(),
			},
		}
		for _, ii := range splits {
			if err := ii.Insert(t.ctx, t.db, boil.Infer()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *transactionImporterService) ofxRespCreditCard(msg ofxgo.Message) error {
	stmt, ok := msg.(*ofxgo.CCStatementResponse)
	if !ok {
		return fmt.Errorf("failed to parse credit card response")
	}

	account, err := t.getCreditCardAccount(stmt.CCAcctFrom)
	if err != nil {
		return err
	}

	currency, err := t.getCurrencyFromOfxCurrencySymbol(stmt.CurDef)
	if err != nil {
		return err
	}

	tx, err := t.db.BeginTx(t.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	imbalancedAccount, err := model.Accounts(qm.Where("name=?", "Imbalanced")).One(t.ctx, tx)
	if err != nil {
		return err
	}

	for _, i := range stmt.BankTranList.Transactions {
		transaction := model.Transaction{
			EntityID:   t.defaultResources.EntityID(),
			Memo:       i.Memo.String(),
			CurrencyID: currency.ID,
			Date:       i.DtPosted.Format("2006-01-02 03:04:05.000"),
		}

		v1Hash := hashers.NewTransactionHashV1()
		transactionHash := v1Hash.Hash(transaction)
		hashType, err := model.TransactionsHashTypes(qm.Where("name=?", v1Hash.String())).One(t.ctx, tx)
		if err != nil {
			return err
		}

		q := []qm.QueryMod{
			qm.Where("hash=?", fmt.Sprintf("%x", transactionHash.Sum(nil))),
			qm.Where("transactions_hash_type_id=?", hashType.ID),
		}
		exists, err := model.TransactionsHashes(q...).Exists(t.ctx, tx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if err := transaction.Insert(t.ctx, tx, boil.Infer()); err != nil {
			return err
		}

		th := model.TransactionsHash{
			TransactionsID:         transaction.ID,
			TransactionsHashTypeID: hashType.ID,
			Hash:                   fmt.Sprintf("%x", transactionHash.Sum(nil)),
		}
		if err := th.Insert(t.ctx, tx, boil.Infer()); err != nil {
			return err
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
			if err := ii.Upsert(t.ctx, tx, true, []string{"transactions_id", "name"}, boil.Whitelist("name"), boil.Infer()); err != nil {
				return err
			}
		}

		splits := []model.Split{
			{
				TransactionsID: transaction.ID,
				AccountID:      account.ID,
				ValueNum:       i.TrnAmt.Num().Int64(),
				ValueDenom:     i.TrnAmt.Denom().Int64(),
			},
			{
				TransactionsID: transaction.ID,
				AccountID:      imbalancedAccount.ID,
				ValueNum:       -i.TrnAmt.Num().Int64(),
				ValueDenom:     i.TrnAmt.Denom().Int64(),
			},
		}
		for _, ii := range splits {
			if err := ii.Insert(t.ctx, tx, boil.Infer()); err != nil {
				return err
			}
		}

	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

type TransactionImporterRequest struct {
	Path     string
	Size     int64
	Contents []byte
	Checksum []byte
}

func NewTransactionImporterRequest(path string, r io.Reader) (TransactionImporterRequest, error) {
	var req TransactionImporterRequest
	var buf bytes.Buffer

	tee := io.TeeReader(r, &buf)

	contents, err := io.ReadAll(tee)
	if err != nil {
		return req, err
	}

	h := sha256.New()
	h.Write(buf.Bytes())

	req = TransactionImporterRequest{
		Path:     path,
		Contents: contents,
		Checksum: h.Sum(nil),
	}

	return req, nil
}

func (t *transactionImporterService) Do(r TransactionImporterRequest) error {
	ofxResp, err := t.ofxParse(bytes.NewReader(r.Contents))
	if err != nil {
		return err
	}

	switch {
	case len(ofxResp.Bank) > 0:
		for _, msg := range ofxResp.Bank {
			if err := t.ofxRespBank(msg); err != nil {
				return err
			}
		}
	case len(ofxResp.CreditCard) > 0:
		for _, msg := range ofxResp.CreditCard {
			if err := t.ofxRespCreditCard(msg); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("ofx file not supported")
	}

	return nil
}

func (t transactionImporterService) createBankAccount(bankAccount ofxgo.BankAcct) (model.Account, error) {
	var account model.Account
	var err error

	tx, err := t.db.BeginTx(t.ctx, nil)
	if err != nil {
		t.logger.Error(err.Error())
		return account, err
	}

	accountType, err := model.AccountTypes(qm.Where("name=?", "Asset")).One(t.ctx, tx)
	if err != nil {
		t.logger.Error(err.Error())
		return account, err
	}

	account = model.Account{
		Name:          fmt.Sprintf("Bank Account %s%s", bankAccount.BankID.String(), bankAccount.AcctID.String()),
		AccountTypeID: accountType.ID,
		EntityID:      t.defaultResources.EntityID(),
		ParentID:      t.defaultResources.GrandParentAccountID(),
	}

	if err := account.Insert(t.ctx, tx, boil.Infer()); err != nil {
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
		if err := i.Insert(t.ctx, tx, boil.Infer()); err != nil {
			return account, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return account, err
	}

	return account, nil
}
