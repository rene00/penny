package types

import (
	"context"
	"database/sql"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/model"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type JSResp struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    any    `json:"data,omitempty"`
}

type Currency struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// isTransactionDeleted accepts a transaction id and returns true if it is marked
// as deleted.
func isTransactionDeleted(ctx context.Context, db *sql.DB, transactionsID int64) (bool, error) {
	var resp bool
	transactionAttribute, err := model.TransactionsAttributes(qm.Where("transactions_id=? AND name=?", transactionsID, "deleted")).One(ctx, db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		return resp, err
	default:
		if transactionAttribute.Value == "true" {
			resp = true
		}
	}
	return resp, nil
}
