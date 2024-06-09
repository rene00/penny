package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"pennyapp/backend/config"
	"pennyapp/backend/model"
	"pennyapp/db/migrations"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const DefaultAccountTypeName string = "AccountType"

type dbService struct {
	ctx  context.Context
	DB   *sql.DB
	conf config.Config

	// string filepath of the sqlite3 dbfile which is set during Start().
	dbFile string

	defaultEntity          *model.Entity
	grandParentAccountType *model.AccountType
}

var db *dbService
var onceDB sync.Once

func DB() *dbService {
	if db == nil {
		onceDB.Do(func() {
			db = &dbService{}
		})
	}
	return db
}

func (d *dbService) Start(ctx context.Context, conf config.Config) {
	var err error
	d.ctx = ctx
	d.conf = conf

	dbFile := conf.DBFile
	if dbFile == "" {
		dataHome, ok := os.LookupEnv("XDG_DATA_HOME")
		if ok {
			dbFile = path.Join(dataHome, "pennyapp", "pennyapp.db")
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			dbFile = path.Join(homeDir, ".local", "share", "pennyapp", "pennyapp.db")
		}
	}

	dbDir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		panic(err)
	}

	err = migrations.DoMigrateDb("sqlite3://" + dbFile)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}

	boil.SetDB(db)
	boil.DebugMode = d.conf.Debug

	d.DB = db

	err = d.initAllData(ctx)
	if err != nil {
		panic(err)
	}

	d.dbFile = dbFile
}

func (d *dbService) initAllData(ctx context.Context) error {

	if err := d.initAccountTypes(ctx); err != nil {
		return err
	}

	if err := d.initEntityTypes(ctx); err != nil {
		return err
	}

	if err := d.initEntity(ctx); err != nil {
		return err
	}

	if err := d.initAccount(ctx); err != nil {
		return err
	}

	if err := d.initCurrency(ctx); err != nil {
		return err
	}

	if err := d.initTransactionsHashType(ctx); err != nil {
		return err
	}

	return nil
}

func (d *dbService) initTransactionsHashType(ctx context.Context) error {
	transactionHashType := &model.TransactionsHashType{Name: "v1"}
	if err := transactionHashType.Insert(ctx, d.DB, boil.Infer()); err != nil {
		switch {
		case isUniqueConstraint(err):
		default:
			return fmt.Errorf("failed to insert transaction hash type: %w", err)
		}
	}
	return nil
}

func (d *dbService) initCurrency(ctx context.Context) error {
	currency := &model.Currency{Name: "Australian Dollar"}
	if err := currency.Insert(ctx, d.DB, boil.Infer()); err != nil {
		switch {
		case isUniqueConstraint(err):
			currency, err = model.Currencies([]qm.QueryMod{
				qm.Where("name=?", currency.Name),
			}...).One(ctx, d.DB)
			if err != nil {
				return fmt.Errorf("failed to insert currency: %w", err)
			}
		default:
			return fmt.Errorf("failed to insert currency: %w", err)
		}
	}

	currencyAttributes := []model.CurrencyAttribute{
		{
			CurrencyID: currency.ID,
			Name:       "code",
			Value:      "AUD",
		},
	}
	for _, i := range currencyAttributes {
		if err := i.Insert(ctx, d.DB, boil.Infer()); err != nil {
			switch {
			case isUniqueConstraint(err):
				continue
			default:
				return fmt.Errorf("failed to insert currency attributes: %w", err)
			}
		}
	}

	return nil
}

func (d *dbService) initEntity(ctx context.Context) error {
	var q []qm.QueryMod

	entityType, err := model.EntityTypes(qm.Where("name=?", "Person")).One(ctx, d.DB)
	if err != nil {
		return fmt.Errorf("failed to find person entity: %w", err)
	}

	entity := &model.Entity{Name: "Default", EntityTypeID: entityType.ID}
	if err := entity.Insert(ctx, d.DB, boil.Infer()); err != nil {
		switch {
		case isUniqueConstraint(err):
			q = []qm.QueryMod{
				qm.Where("name=?", "Default"),
				qm.Where("entity_type_id=?", entityType.ID),
			}
			entity, err = model.Entities(q...).One(ctx, d.DB)
			if err != nil {
				return fmt.Errorf("failed to insert default entity: %w", err)
			}
		default:
			return fmt.Errorf("failed to insert entity: %w", err)
		}
	}
	d.defaultEntity = entity

	return nil
}

func (d *dbService) getDefaultEntity() (*model.Entity, error) {
	if d.defaultEntity == nil {
		return nil, fmt.Errorf("no default entity")
	}
	return d.defaultEntity, nil
}

func (d *dbService) getGrantParentAccountType() (*model.AccountType, error) {
	if d.grandParentAccountType == nil {
		return nil, fmt.Errorf("no grand parent account type")
	}
	return d.grandParentAccountType, nil
}

func (d *dbService) initAccount(ctx context.Context) error {
	defaultEntity, err := d.getDefaultEntity()
	if err != nil {
		return err
	}
	_, err = model.Accounts(qm.Where("name=? AND parent_id=0 AND entity_id=?", "Account", defaultEntity.ID)).One(ctx, d.DB)
	if err != nil {
		switch {
		case isNoRowsFound(err):
			grandParentAccountType, err := d.getGrantParentAccountType()
			if err != nil {
				return err
			}

			grandParentAccount := &model.Account{
				Name:          "Account",
				EntityID:      defaultEntity.ID,
				AccountTypeID: grandParentAccountType.ID,
				ParentID:      int64(0),
			}
			if err := grandParentAccount.Insert(ctx, d.DB, boil.Infer()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("failed to insert grand parent account: %w", err)
		}
	}

	return nil
}

func (d *dbService) initEntityTypes(ctx context.Context) error {
	entityTypes := []model.EntityType{
		{Name: "Person"},
	}

	for _, entityType := range entityTypes {
		if err := entityType.Insert(ctx, d.DB, boil.Infer()); err != nil {
			switch {
			case isUniqueConstraint(err):
				continue
			default:
				return fmt.Errorf("failed to insert entity type: %w", err)
			}
		}
	}

	return nil
}

func (d *dbService) initAccountTypes(ctx context.Context) error {

	grandParentAccountType, err := model.AccountTypes(qm.Where("name=?", DefaultAccountTypeName)).One(ctx, d.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			grandParentAccountType = &model.AccountType{Name: DefaultAccountTypeName}
			if err := grandParentAccountType.Insert(ctx, d.DB, boil.Infer()); err != nil {
				return err
			}
		}
	}
	d.grandParentAccountType = grandParentAccountType

	parentAccountTypes := []model.AccountType{
		{Name: "Liabilities", ParentID: grandParentAccountType.ID},
		{Name: "Assets", ParentID: grandParentAccountType.ID},
		{Name: "Revenue", ParentID: grandParentAccountType.ID},
		{Name: "Expenses", ParentID: grandParentAccountType.ID},
		{Name: "Equity", ParentID: grandParentAccountType.ID},
	}

	for _, parentAccountType := range parentAccountTypes {
		if err := parentAccountType.Insert(ctx, d.DB, boil.Infer()); err != nil {
			switch {
			case isUniqueConstraint(err):
				continue
			default:
				return fmt.Errorf("failed to insert parent account type: %w", err)
			}
		}
	}

	type parentAccountTypeName string

	childAccountTypes := map[parentAccountTypeName][]model.AccountType{
		parentAccountTypeName("Liabilities"): {
			{Name: "Liability"},
		},
		parentAccountTypeName("Expenses"): {
			{Name: "Expense"},
		},
		parentAccountTypeName("Revenue"): {
			{Name: "Income"},
		},
		parentAccountTypeName("Assets"): {
			{Name: "Asset"},
		},
	}

	for k, v := range childAccountTypes {
		parentAccountType, err := model.AccountTypes(qm.Where("name=? AND parent_id=?", string(k), grandParentAccountType.ID)).One(ctx, d.DB)
		if err != nil {
			return fmt.Errorf("failed to find parent account type: %w", err)
		}
		for _, accountType := range v {
			accountType.ParentID = parentAccountType.ID
			if err := accountType.Insert(ctx, d.DB, boil.Infer()); err != nil {
				switch {
				case isUniqueConstraint(err):
					continue
				default:
					return fmt.Errorf("failed to insert child account type: %w", err)
				}
			}
		}
	}

	return nil
}

func isNoRowsFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
