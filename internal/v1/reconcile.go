package v1

import (
	"context"
	"database/sql"
	"fmt"
)

func ReconcileTransaction(ctx context.Context, db *sql.DB, rt, t Transaction) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if rt.Debit != 0 && rt.Debit != t.Debit {
		stmt, err := tx.PrepareContext(ctx, "UPDATE tx SET debit = ? WHERE id = ? LIMIT 1")
		if err != nil {
			return err
		}

		_, err = stmt.ExecContext(ctx, rt.Debit, t.ID)
		if err != nil {
			return err
		}
		fmt.Printf("debug: update debit %g\n", rt.Debit)
	}

	if rt.Credit != 0 && rt.Credit != t.Credit {
		stmt, err := tx.PrepareContext(ctx, "UPDATE tx SET credit = ? WHERE id = ? LIMIT 1")
		if err != nil {
			return err
		}

		_, err = stmt.ExecContext(ctx, rt.Credit, t.ID)
		if err != nil {
			return err
		}
		fmt.Printf("debug: update credit %g\n", rt.Credit)
	}

	for _, s := range rt.SplitTransactions {
		stmt, err := tx.PrepareContext(ctx, "INSERT INTO tx (id, user_id, date, debit, credit, memo, bankaccount_id, parent_id, account_id) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}
		_, err = stmt.ExecContext(ctx, t.UserID, t.Date, s.Debit, s.Credit, s.Memo, t.BankAccountID, t.ID, s.AccountID)
		if err != nil {
			return err
		}

		fmt.Printf("%#v\n", s)
	}

	return tx.Commit()
}
