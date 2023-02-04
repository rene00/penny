package seed

import (
	"context"
	"penny/internal/handler"
	"penny/internal/model"

	"gorm.io/gorm"
)

type Seed struct {
	db        *gorm.DB
	dbHandler handler.DB
}

func NewSeed(db *gorm.DB) Seed {
	return Seed{db: db, dbHandler: handler.NewDBHandler()}
}

func (s Seed) FinanceInstitutionType(ctx context.Context) error {
	financeInstitutionTypes := []model.FinanceInstitutionType{
		{Name: "Bank"},
	}
	for _, i := range financeInstitutionTypes {
		if err := s.db.WithContext(ctx).Create(&i).Error; err != nil && !s.dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

func (s Seed) ParentAccountType(ctx context.Context) error {
	parentAccountTypes := []model.ParentAccountType{
		{Name: "Finance Institution"},
		{Name: "Cash"},
	}

	for _, i := range parentAccountTypes {
		if err := s.db.WithContext(ctx).Create(&i).Error; err != nil && !s.dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

func (s Seed) AccountType(ctx context.Context) error {
	accountTypes := []model.AccountType{
		{
			Name:              "Checking",
			ParentAccountType: model.ParentAccountType{Name: "Finance Institution"},
		},
		{
			Name:              "Savings",
			ParentAccountType: model.ParentAccountType{Name: "Finance Institution"},
		},
		{
			Name:              "Home Loan",
			ParentAccountType: model.ParentAccountType{Name: "Finance Institution"},
		},
		{
			Name:              "Credit Card",
			ParentAccountType: model.ParentAccountType{Name: "Finance Institution"},
		},
		{
			Name:              "Cash",
			ParentAccountType: model.ParentAccountType{Name: "Cash"},
		},
	}

	for _, i := range accountTypes {
		tx := s.db.WithContext(ctx)
		if err := tx.First(&i.ParentAccountType).Error; err != nil {
			return err
		}
		if err := tx.Create(&i).Error; err != nil && !s.dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

func (s Seed) FinanceInstitution(ctx context.Context) error {
	financeInstitution := []model.FinanceInstitution{
		{
			Name:                   "Commonwealth Bank of Australia",
			FinanceInstitutionType: model.FinanceInstitutionType{Name: "Bank"},
		},
		{
			Name:                   "Westpac",
			FinanceInstitutionType: model.FinanceInstitutionType{Name: "Bank"},
		},
		{
			Name:                   "PayPal",
			FinanceInstitutionType: model.FinanceInstitutionType{Name: "Bank"},
		},
		{
			Name:                   "Macquarie Bank",
			FinanceInstitutionType: model.FinanceInstitutionType{Name: "Bank"},
		},
	}

	for _, i := range financeInstitution {
		if err := s.db.WithContext(ctx).Create(&i).Error; err != nil && !s.dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

func (s Seed) Account(ctx context.Context) error {
	accounts := []model.Account{
		{
			Name:        "Cash",
			Entity:      model.Entity{Name: "Personal"},
			AccountType: model.AccountType{Name: "Cash"},
		},
	}
	for _, i := range accounts {
		if err := s.db.WithContext(ctx).Create(&i).Error; err != nil && !s.dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}
