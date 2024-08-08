package services

import (
	"context"
	"database/sql"
	"fmt"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"regexp"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type accountMatchService struct {
	ctx                  context.Context
	db                   *sql.DB
	logger               *logwrap.LogWrap
	defaultResources     defaultresources.DefaultResources
	AccountMatchRequestC chan AccountMatchRequest
}

type AccountMatchRequest struct{}

var accountMatch *accountMatchService

func AccountMatch() *accountMatchService {
	return &accountMatchService{}
}

func (a *accountMatchService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap, defaultResources defaultresources.DefaultResources) {
	a.ctx = ctx
	a.db = db
	a.logger = logger
	a.defaultResources = defaultResources
	a.AccountMatchRequestC = make(chan AccountMatchRequest)

	go func() { a.matcher(a.ctx, a.db, a.logger) }()
}

func (a *accountMatchService) matcher(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	var q []qm.QueryMod

	for range a.AccountMatchRequestC {
		logger.Info(fmt.Sprintf("received new account match request"))

		accountMatches, err := model.AccountMatches().All(ctx, db)
		if err != nil {
			logger.Error("Failed to find Account Matches")
		}

		q = []qm.QueryMod{
			qm.Load("Splits"),
			qm.Load("Splits.Account"),
		}
		transactions, err := model.Transactions(q...).All(ctx, db)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		for _, accountMatch := range accountMatches {
			attr, err := model.AccountMatchAttributes(qm.Where("account_match_id=? AND name=?", accountMatch.ID, "deleted")).One(ctx, db)
			switch {
			case dberror.IsNoRowsFound(err):
			case err != nil:
				logger.Error(err.Error())
				continue
			default:
				if attr.Value == "true" {
					continue
				}
			}

			accountMatchRegexes, err := model.AccountMatchRegexes(qm.Where("account_match_id=?", accountMatch.ID)).All(ctx, db)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			for _, transaction := range transactions {
				transactionDeleted, err := model.TransactionsAttributes(qm.Where("transactions_id=? AND name=? AND value=?", transaction.ID, "deleted", "true")).Exists(ctx, db)
				if err != nil {
					logger.Error(err.Error())
					continue
				}
				if transactionDeleted {
					continue
				}

				for _, split := range transaction.R.Splits {
					if accountMatch.SourceAccountID == split.R.Account.ID {
						exists, err := model.Splits(qm.Where("transactions_id=? AND id!=?", split.TransactionsID, split.ID)).Exists(ctx, db)
						if err != nil {
							logger.Error(err.Error())
							continue
						}
						if exists {
							continue
						}

						for _, regex := range accountMatchRegexes {
							r, err := regexp.Compile(fmt.Sprintf("(?i)%s", regex.Regex))
							if err != nil {
								logger.Error(fmt.Sprintf("regexp compile %d: %s", regex.ID, err.Error()))
								continue
							}

							result := r.FindStringSubmatch(transaction.Memo)

							valueNum := split.ValueNum
							if split.ValueNum < 0 {
								valueNum = split.ValueNum * -1
							}

							if len(result) > 0 {
								s := model.Split{
									TransactionsID: transaction.ID,
									AccountID:      accountMatch.DestinationAccountID,
									ValueNum:       valueNum,
									ValueDenom:     split.ValueDenom,
								}
								if err := s.Insert(ctx, db, boil.Infer()); err != nil {
									logger.Error(err.Error())
									continue
								}
							}
						}
					}
				}
			}
		}
	}
}

func (a *accountMatchService) SendAccountMatchRequest() types.AccountMatchesResponse {
	var resp types.AccountMatchesResponse
	a.AccountMatchRequestC <- AccountMatchRequest{}
	resp.Success = true
	resp.Msg = "Request sent"
	return resp
}

func (a *accountMatchService) GetAccountMatches() types.AccountMatchesResponse {
	resp := types.AccountMatchesResponse{}

	q := []qm.QueryMod{
		qm.InnerJoin("account source_account on account_match.source_account_id = source_account.id"),
		qm.InnerJoin("account destination_account on account_match.destination_account_id = destination_account.id"),
	}

	accountMatches, err := model.AccountMatches(q...).All(a.ctx, a.db)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Data = []types.AccountMatch{}
	for _, i := range accountMatches {
		am, err := types.NewAccountMatch(a.ctx, a.db, i)
		if err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}
		if am.Deleted {
			continue
		}
		resp.Data = append(resp.Data, am)
	}
	resp.Success = true
	return resp
}

func (a *accountMatchService) createAccountMatch(name, description string, sourceAccountID, destinationAccountID int64) (*model.AccountMatch, error) {
	am := &model.AccountMatch{
		Name:                 name,
		SourceAccountID:      sourceAccountID,
		DestinationAccountID: destinationAccountID,
	}

	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		return am, err
	}
	defer tx.Rollback()

	if err := am.Upsert(a.ctx, tx, true, []string{"name", "source_account_id", "destination_account_id"}, boil.Whitelist("id", "name", "source_account_id", "destination_account_id"), boil.Infer()); err != nil {
		return am, err
	}

	if len(description) > 0 {
		attribute := model.AccountMatchAttribute{
			AccountMatchID: am.ID,
			Name:           "description",
			Value:          description,
		}
		if err := attribute.Insert(a.ctx, tx, boil.Infer()); err != nil {
			return am, err
		}
	}

	if err := tx.Commit(); err != nil {
		return am, err
	}

	return am, nil
}

func (a *accountMatchService) CreateAccountMatchRegex(id int64, regex string) types.AccountMatchRegexResponse {
	var resp types.AccountMatchRegexResponse
	am, err := model.FindAccountMatch(a.ctx, a.db, id)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	r := model.AccountMatchRegex{
		AccountMatchID: am.ID,
		Regex:          regex,
	}

	if err := r.Insert(a.ctx, a.db, boil.Infer()); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}

func (a *accountMatchService) CreateAccountMatch(name, description string, sourceAccountID, destinationAccountID int64) types.AccountMatchResponse {
	var resp types.AccountMatchResponse

	accountMatch, err := a.createAccountMatch(name, description, sourceAccountID, destinationAccountID)
	switch {
	case dberror.IsUniqueConstraint(err):
		resp.Msg = "Account Match already exists"
		a.logger.Error(resp.Msg)
		return resp
	case err != nil:
		resp.Msg = fmt.Sprintf("Failed to create Account Match: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Data, err = types.NewAccountMatch(a.ctx, a.db, accountMatch)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Account Match created"
	return resp
}

func (a *accountMatchService) GetAccountMatch(id int64) types.AccountMatchResponse {
	var resp types.AccountMatchResponse
	var err error

	q := []qm.QueryMod{
		qm.Where("account_match.id=?", id),
		qm.InnerJoin("account source_account on account_match.source_account_id = source_account.id"),
		qm.InnerJoin("account destination_account on account_match.destination_account_id = destination_account.id"),
	}

	am, err := model.AccountMatches(q...).One(a.ctx, a.db)
	switch {
	case dberror.IsNoRowsFound(err):
		resp.Msg = "Account Match not found"
		a.logger.Error(resp.Msg)
		return resp
	case err != nil:
		resp.Msg = fmt.Sprintf("Failed to find Account Match: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Data, err = types.NewAccountMatch(a.ctx, a.db, am)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true

	return resp
}

func (a *accountMatchService) UpdateAccountMatch(id, sourceAccountID, destinationAccountID int64, name, description string) types.AccountMatchResponse {
	var resp types.AccountMatchResponse
	var err error

	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	am, err := model.FindAccountMatch(a.ctx, tx, id)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	am.Name = name
	am.SourceAccountID = sourceAccountID
	am.DestinationAccountID = destinationAccountID

	if _, err = am.Update(a.ctx, tx, boil.Whitelist("source_account_id", "destination_account_id", "name")); err != nil {
		resp.Msg = err.Error()
		return resp
	}

	attr := model.AccountMatchAttribute{
		ID:             0,
		AccountMatchID: am.ID,
		Name:           "description",
		Value:          description,
	}
	if err = attr.Upsert(a.ctx, tx, true, []string{"account_match_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
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
	resp.Msg = "Account Match updated"

	_ = a.SendAccountMatchRequest()
	return resp
}

func (a *accountMatchService) UpdateAccountMatchRegex(id int64, regex string) types.AccountMatchRegexResponse {
	var resp types.AccountMatchRegexResponse
	accountMatchRegex, err := model.FindAccountMatchRegex(a.ctx, a.db, id)
	if err != nil {
		resp.Msg = "Failed to find Account Match Regex"
		a.logger.Error(resp.Msg)
		return resp
	}

	accountMatchRegex.Regex = regex
	if _, err := accountMatchRegex.Update(a.ctx, a.db, boil.Infer()); err != nil {
		resp.Msg = "Failed to update Account Match Regex"
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Account Match Filter updated"
	_ = a.SendAccountMatchRequest()
	return resp
}

func (a *accountMatchService) DeleteAccountMatchRegex(id int64) types.AccountMatchRegexResponse {
	var resp types.AccountMatchRegexResponse
	if _, err := model.AccountMatchRegexes(qm.Where("id=?", id)).DeleteAll(a.ctx, a.db); err != nil {
		resp.Msg = "Failed to delete Account Match Regex"
		a.logger.Error(resp.Msg)
		return resp
	}
	resp.Success = true
	resp.Msg = "Account Match Filter deleted"
	return resp
}

func (a *accountMatchService) DeleteAccountMatch(id int64) types.AccountMatchResponse {
	var resp types.AccountMatchResponse
	attr := model.AccountMatchAttribute{
		AccountMatchID: id,
		Name:           "deleted",
		Value:          "true",
	}
	if err := attr.Upsert(a.ctx, a.db, true, []string{"account_match_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	resp.Success = true
	resp.Msg = "Account Match deleted"
	return resp
}

func (a *accountMatchService) UndeleteAccountMatch(id int64) types.AccountMatchResponse {
	var resp types.AccountMatchResponse
	attr := model.AccountMatchAttribute{
		AccountMatchID: id,
		Name:           "deleted",
		Value:          "false",
	}
	if err := attr.Upsert(a.ctx, a.db, true, []string{"account_match_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		return resp
	}
	resp.Success = true
	resp.Msg = "Account Match undeleted"
	return resp
}
