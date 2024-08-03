package types

import (
	"context"
	"database/sql"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/model"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AccountMatchResponse struct {
	Success bool         `json:"success"`
	Msg     string       `json:"msg"`
	Data    AccountMatch `json:"data"`
}

type AccountMatch struct {
	ID                 int64               `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Deleted            bool                `json:"deleted"`
	SourceAccount      Account             `json:"source_account"`
	DestinationAccount Account             `json:"destination_account"`
	Regexes            []AccountMatchRegex `json:"regexes"`
	modelAccountMatch  *model.AccountMatch
}

type AccountMatchesResponse struct {
	Success bool           `json:"success"`
	Msg     string         `json:"msg"`
	Data    []AccountMatch `json:"data"`
}

type AccountMatchRegexResponse struct {
	Success bool              `json:"success"`
	Msg     string            `json:"msg"`
	Data    AccountMatchRegex `json:"data"`
}

type AccountMatchRegex struct {
	ID    int64  `json:"id"`
	Regex string `json:"regex"`
}

func NewAccountMatch(ctx context.Context, db *sql.DB, a *model.AccountMatch) (AccountMatch, error) {
	var err error
	am := AccountMatch{
		ID:                a.ID,
		Name:              a.Name,
		modelAccountMatch: a,
	}

	if err = am.setSourceAccount(ctx, db); err != nil {
		return am, err
	}
	if err = am.setDestinationAccount(ctx, db); err != nil {
		return am, err
	}
	if err = am.setRegexes(ctx, db); err != nil {
		return am, err
	}
	if err = am.setDescription(ctx, db); err != nil {
		return am, err
	}
	if err = am.setDeleted(ctx, db); err != nil {
		return am, err
	}
	return am, nil
}

func (a *AccountMatch) setRegexes(ctx context.Context, db *sql.DB) error {
	regexes, err := model.AccountMatchRegexes(qm.Where("account_match_id=?", a.modelAccountMatch.ID)).All(ctx, db)
	if err != nil {
		return err
	}
	a.Regexes = []AccountMatchRegex{}
	for _, regex := range regexes {
		a.Regexes = append(a.Regexes, AccountMatchRegex{ID: regex.ID, Regex: regex.Regex})
	}
	return nil
}

func (a *AccountMatch) setDeleted(ctx context.Context, db *sql.DB) error {
	attr, err := model.AccountMatchAttributes(qm.Where("account_match_id=? AND name=?", a.modelAccountMatch.ID, "deleted")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		if attr.Value == "true" {
			a.Deleted = true
		}
	}
	return nil
}

func (a *AccountMatch) setSourceAccount(ctx context.Context, db *sql.DB) error {
	sourceAccount, err := a.modelAccountMatch.SourceAccount().One(ctx, db)
	if err != nil {
		return err
	}
	a.SourceAccount = Account{
		ID:   sourceAccount.ID,
		Name: sourceAccount.Name,
	}
	return nil
}

func (a *AccountMatch) setDestinationAccount(ctx context.Context, db *sql.DB) error {
	destinationAccount, err := a.modelAccountMatch.DestinationAccount().One(ctx, db)
	if err != nil {
		return err
	}
	a.DestinationAccount = Account{
		ID:   destinationAccount.ID,
		Name: destinationAccount.Name,
	}
	return nil
}

func (a *AccountMatch) setDescription(ctx context.Context, db *sql.DB) error {
	attr, err := model.AccountMatchAttributes(qm.Where("account_match_id=? AND name=?", a.modelAccountMatch.ID, "description")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		a.Description = attr.Value
	}
	return nil
}
