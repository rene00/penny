package services

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const DefaultCurrencyDenominator int64 = 100

type splitsService struct {
	ctx    context.Context
	db     *sql.DB
	logger *logwrap.LogWrap
}

var splits *splitsService

func Splits() *splitsService {
	splits = &splitsService{}
	return splits
}

func (s *splitsService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	s.ctx = ctx
	s.db = db
	s.logger = logger
	return
}

func (s *splitsService) getSplit(splitID int64) (*model.Split, error) {
	q := []qm.QueryMod{
		qm.Where("splits.id=?", splitID),
		qm.Load("Account.AccountType"),
	}
	return model.Splits(q...).One(s.ctx, s.db)
}

func (s *splitsService) deleteSplit(split *model.Split) error {
	splitsAttribute := model.SplitsAttribute{
		SplitsID: split.ID,
		Name:     "deleted",
		Value:    "true",
	}
	return splitsAttribute.Upsert(s.ctx, s.db, true, []string{"splits_id", "name"}, boil.Whitelist("name"), boil.Infer())
}

func (s *splitsService) undeleteSplit(split *model.Split) error {
	_, err := model.SplitsAttributes(qm.Where("splits_id=? AND name=?", split.ID, "deleted")).DeleteAll(s.ctx, s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *splitsService) updateSplit(splitID, accountID int64, amount float64) (*model.Split, error) {
	var split *model.Split
	var err error

	split, err = s.getSplit(splitID)
	if err != nil {
		return nil, err
	}

	split.AccountID = accountID

	r := big.NewRat(int64(amount*float64(DefaultCurrencyDenominator)), DefaultCurrencyDenominator)
	split.ValueNum = r.Num().Int64()
	split.ValueDenom = r.Denom().Int64()

	if _, err = split.Update(s.ctx, s.db, boil.Whitelist("transactions_id", "account_id", "value_num", "value_denom")); err != nil {
		return nil, err
	}

	return split, nil
}

func (s *splitsService) UndeleteSplit(id int64) types.SplitResponse {
	var resp types.SplitResponse

	split, err := s.getSplit(id)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to find split: %s", err)
		return resp
	}

	if err := s.undeleteSplit(split); err != nil {
		resp.Msg = fmt.Sprintf("Failed to undelete split: %s", err)
		s.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Split undeleted"
	resp.Success = true

	return resp
}

func (s *splitsService) DeleteSplit(id int64) types.SplitResponse {
	var resp types.SplitResponse

	split, err := s.getSplit(id)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to find split: %s", err)
		return resp
	}

	if err := s.deleteSplit(split); err != nil {
		resp.Msg = fmt.Sprintf("Failed to delete split: %s", err)
		s.logger.Error(resp.Msg)
		return resp
	}

	resp.Msg = "Split deleted"
	resp.Success = true

	return resp
}

func (s *splitsService) UpdateSplit(splitID, accountID int64, amount float64) types.SplitResponse {
	var resp types.SplitResponse
	var split *model.Split
	var err error

	split, err = s.updateSplit(splitID, accountID, amount)
	if err != nil {
		resp.Msg = err.Error()
		s.logger.Error(resp.Msg)
		return resp
	}

	resp = types.NewSplitResponse(*split)
	resp.Success = true
	return resp
}

func (s *splitsService) CreateSplit(transactionID, accountID int64, amount float64) types.JSResp {
	var resp types.JSResp

	if transactionID == 0 {
		resp.Msg = "Transaction must not be empty"
		s.logger.Error(resp.Msg)
		return resp
	}

	if accountID == 0 {
		resp.Msg = "Account must not be empty"
		s.logger.Error(resp.Msg)
		return resp
	}

	r := big.NewRat(int64(amount*float64(DefaultCurrencyDenominator)), DefaultCurrencyDenominator)

	split := model.Split{
		TransactionsID: transactionID,
		AccountID:      accountID,
		ValueNum:       r.Num().Int64(),
		ValueDenom:     r.Denom().Int64(),
	}

	if err := split.Insert(s.ctx, s.db, boil.Infer()); err != nil {
		resp.Msg = fmt.Sprintf("Failed to insert split: %s", err.Error())
		s.logger.Error(resp.Msg)
		return resp
	}
	resp.Msg = "Created split"
	resp.Success = true

	return resp
}
