package types

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/model"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ReportProfitLossResponse struct {
	Success bool             `json:"success"`
	Msg     string           `json:"msg"`
	Data    ReportProfitLoss `json:"data"`
}

type ReportProfitLoss struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`

	// Net is the product of "Income Total Amount - Expense Total Amount"
	Net string `json:"net"`

	Income  []Account `json:"income"`
	Expense []Account `json:"expense"`
}

func (r *ReportProfitLoss) endDate(ctx context.Context, db *sql.DB) error {
	q := []qm.QueryMod{
		qm.Select("transactions.date as start_date"),
		qm.Where("transactions_attributes.name!=? AND transactions_attributes.value!=?", "deleted", "true"),
		qm.OrderBy("transactions.date DESC"),
		qm.Limit(1),
		qm.InnerJoin("transactions_attributes on transactions_attributes.transactions_id = transactions.id"),
		qm.Load("TransactionsAttributes"),
	}

	endTransaction, err := model.Transactions(q...).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
		return nil
	case err != nil:
		return err
	}
	r.EndDate = fmt.Sprintf("end:%s", endTransaction.Date)
	return nil
}

func (r *ReportProfitLoss) expense(ctx context.Context, db *sql.DB) error {
	expenseAccountType, err := model.AccountTypes(qm.Where("name=?", "Expense")).One(ctx, db)
	if err != nil {
		return err
	}

	accounts, err := model.Accounts(qm.Where("account_type_id=?", expenseAccountType.ID)).All(ctx, db)
	if err != nil {
		return err
	}

	r.Expense = []Account{}

	for _, a := range accounts {
		account, err := NewAccount(ctx, db, a)
		if err != nil {
			return err
		}
		if account.Deleted {
			continue
		}
		r.Expense = append(r.Expense, account)
	}

	return nil
}

func (r *ReportProfitLoss) income(ctx context.Context, db *sql.DB) error {
	incomeAccountType, err := model.AccountTypes(qm.Where("name=?", "Income")).One(ctx, db)
	if err != nil {
		return err
	}

	accounts, err := model.Accounts(qm.Where("account_type_id=?", incomeAccountType.ID)).All(ctx, db)
	if err != nil {
		return err
	}

	r.Income = []Account{}

	for _, a := range accounts {
		account, err := NewAccount(ctx, db, a, []AccountOptFunc{WithAccountOptsAmountAlwaysPosiive(true)}...)
		if err != nil {
			return err
		}
		if account.Deleted {
			continue
		}
		r.Income = append(r.Income, account)
	}

	return nil
}

func (r *ReportProfitLoss) net(ctx context.Context, db *sql.DB) error {

	totalIncomeAmount := float64(0)
	totalExpenseAmount := float64(0)

	for _, account := range r.Income {
		splits, err := model.Splits(qm.Where("account_id=?", account.ID)).All(ctx, db)
		if err != nil {
			return err
		}

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
			totalIncomeAmount = totalIncomeAmount + f
		}
	}

	for _, account := range r.Expense {
		splits, err := model.Splits(qm.Where("account_id=?", account.ID)).All(ctx, db)
		if err != nil {
			return err
		}

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
			totalExpenseAmount = totalExpenseAmount + f
		}
	}

	r.Net = decimal.NewFromFloat(totalIncomeAmount - totalExpenseAmount).StringFixed(2)

	return nil
}

func (r *ReportProfitLoss) startDate(ctx context.Context, db *sql.DB) error {
	q := []qm.QueryMod{
		qm.Select("transactions.date as start_date"),
		qm.Where("transactions_attributes.name!=? AND transactions_attributes.value!=?", "deleted", "true"),
		qm.OrderBy("transactions.date ASC"),
		qm.Limit(1),
		qm.InnerJoin("transactions_attributes on transactions_attributes.transactions_id = transactions.id"),
		qm.Load("TransactionsAttributes"),
	}

	startTransaction, err := model.Transactions(q...).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
		return nil
	case err != nil:
		return err
	}
	r.StartDate = fmt.Sprintf("start:%s", startTransaction.Date)
	return nil
}

func NewReportProfitLoss(ctx context.Context, db *sql.DB) (ReportProfitLoss, error) {
	var err error
	r := ReportProfitLoss{}
	if err = r.startDate(ctx, db); err != nil {
		return r, err
	}
	if err = r.endDate(ctx, db); err != nil {
		return r, err
	}
	if err = r.income(ctx, db); err != nil {
		return r, err
	}
	if err = r.expense(ctx, db); err != nil {
		return r, err
	}
	if err = r.net(ctx, db); err != nil {
		return r, err
	}
	return r, nil
}
