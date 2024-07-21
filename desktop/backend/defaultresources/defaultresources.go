package defaultresources

import (
	"context"
	"database/sql"
	"fmt"
	"pennyapp/backend/model"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DefaultResources struct {
	grandParentAccount          model.Account
	entity                      model.Entity
	transactionImporterStatuses map[string]model.TransactionImporterStatus
}

func NewDefaultResources(ctx context.Context, db *sql.DB) (DefaultResources, error) {
	var err error

	d := DefaultResources{}

	if err = d.setGrandParentAccount(ctx, db); err != nil {
		return d, err
	}

	if err = d.setEntity(ctx, db); err != nil {
		return d, err
	}

	if err = d.setTransactionImporterStatuses(ctx, db); err != nil {
		return d, err
	}

	return d, nil
}

func (d DefaultResources) GrandParentAccountID() int64 {
	return d.grandParentAccount.ID
}

func (d DefaultResources) GrandParentAccount() model.Account {
	return d.grandParentAccount
}

func (d DefaultResources) EntityID() int64 {
	return d.entity.ID
}

func (d DefaultResources) Entity() model.Entity {
	return d.entity
}

func (d DefaultResources) TransactionImporterStatusPending() model.TransactionImporterStatus {
	return d.transactionImporterStatuses["pending"]
}

func (d DefaultResources) TransactionImporterStatusFailed() model.TransactionImporterStatus {
	return d.transactionImporterStatuses["failed"]
}

func (d DefaultResources) TransactionImporterStatusCompleted() model.TransactionImporterStatus {
	return d.transactionImporterStatuses["completed"]
}

func (d *DefaultResources) setTransactionImporterStatuses(ctx context.Context, db *sql.DB) error {
	statuses, err := model.TransactionImporterStatuses().All(ctx, db)
	if err != nil {
		return err
	}
	d.transactionImporterStatuses = map[string]model.TransactionImporterStatus{}
	for _, status := range statuses {
		d.transactionImporterStatuses[status.Name] = *status
	}
	return nil
}

func (d *DefaultResources) setEntity(ctx context.Context, db *sql.DB) error {
	entity, err := model.Entities(qm.Where("name=?", "Default")).One(ctx, db)
	switch {
	case err != nil:
		return err
	case entity == nil:
		return fmt.Errorf("failed to find default entity")
	}
	d.entity = *entity
	return nil
}

func (d *DefaultResources) setGrandParentAccount(ctx context.Context, db *sql.DB) error {
	q := []qm.QueryMod{
		qm.Where("name=?", "Account"),
		qm.Where("parent_id=0"),
	}
	account, err := model.Accounts(q...).One(ctx, db)
	switch {
	case err != nil:
		return err
	case account == nil:
		return fmt.Errorf("failed to find default grand parent account")
	}
	d.grandParentAccount = *account
	return nil
}
