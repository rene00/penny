package types

import (
	"context"
	"database/sql"
	"math/big"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/model"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TagResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    Tag    `json:"data"`
}

type TagsResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    []Tag  `json:"data"`
}

type Tag struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Deleted     bool       `json:"deleted"`
	Regexes     []TagRegex `json:"regexes"`
	Amount      string     `json:"amount"`
}

type TagRegexResponse struct {
	Success bool     `json:"success"`
	Msg     string   `json:"msg"`
	Data    TagRegex `json:"data"`
}

type TagRegex struct {
	ID    int64  `json:"id"`
	Regex string `json:"regex"`
}

func NewTag(ctx context.Context, db *sql.DB, t *model.Tag) (Tag, error) {
	var err error
	tag := Tag{
		ID:   t.ID,
		Name: t.Name,
	}

	if err = tag.setDescription(ctx, db, t); err != nil {
		return tag, err
	}

	if err = tag.setDeleted(ctx, db, t); err != nil {
		return tag, err
	}

	if err = tag.setRegexes(ctx, db, t); err != nil {
		return tag, err
	}

	if err = tag.setAmount(ctx, db, t); err != nil {
		return tag, err
	}

	return tag, nil
}

func (t *Tag) setDescription(ctx context.Context, db *sql.DB, tag *model.Tag) error {
	attr, err := model.TagAttributes(qm.Where("tag_id=? AND name=?", tag.ID, "description")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		t.Description = attr.Value
	}
	return nil
}

func (t *Tag) setDeleted(ctx context.Context, db *sql.DB, tag *model.Tag) error {
	attr, err := model.TagAttributes(qm.Where("tag_id=? AND name=?", tag.ID, "deleted")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return err
	default:
		if attr.Value == "true" {
			t.Deleted = true
		}
	}
	return nil
}

func (t *Tag) setRegexes(ctx context.Context, db *sql.DB, tag *model.Tag) error {
	regexes, err := model.TagRegexes(qm.Where("tag_id=?", tag.ID)).All(ctx, db)
	if err != nil {
		return err
	}
	t.Regexes = []TagRegex{}
	for _, regex := range regexes {
		t.Regexes = append(t.Regexes, TagRegex{ID: regex.ID, Regex: regex.Regex})
	}
	return nil
}

func (t *Tag) setAmount(ctx context.Context, db *sql.DB, tag *model.Tag) error {
	q := []qm.QueryMod{
		qm.Where("tag_transactions.tag_id=?", tag.ID),
		qm.InnerJoin("transactions on tag_transactions.transactions_id = transactions.id"),
		qm.Load("Transaction"),
		qm.Load("Transaction.Splits"),
		qm.Load("Transaction.Splits.Account"),
	}
	tagTransactions, err := model.TagTransactions(q...).All(ctx, db)
	if err != nil {
		return err
	}

	amount := float64(0)
	for _, tagTransaction := range tagTransactions {
		if tagTransaction.R == nil && tagTransaction.R.Transaction == nil {
			continue
		}
		transaction := tagTransaction.R.Transaction

		transactionDeleted, err := isTransactionDeleted(ctx, db, transaction.ID)
		if err != nil {
			return err
		}
		if transactionDeleted {
			continue
		}

		for _, split := range transaction.R.Splits {
			q = []qm.QueryMod{
				qm.Where("account.id=?", split.AccountID),
				qm.InnerJoin("account_type on account_type.id = account.account_type_id"),
				qm.Load("AccountType"),
			}
			account, err := model.Accounts(q...).One(ctx, db)
			if err != nil {
				return err
			}

			accountType, err := account.AccountType().One(ctx, db)
			if err != nil {
				return err
			}

			// Only calculate amount for splits which have an account type of
			// expense. There is a FEAT that improves Tag support by tagging
			// transactions by account type which should remove this hack.
			switch strings.ToLower(accountType.Name) {
			case "expense":
				r := big.NewRat(split.ValueNum, split.ValueDenom)
				f, _ := r.Float64()
				amount = amount + f
			default:
			}

		}
	}
	t.Amount = decimal.NewFromFloat(amount).StringFixed(2)
	return nil
}
