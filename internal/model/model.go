package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
}

type EntityType struct {
	ID   int
	Name string `gorm:"unique;not null"`
}

type Entity struct {
	gorm.Model
	Name         string `gorm:"unique;not null"`
	UserID       int
	User         User
	EntityTypeID int
	EntityType   EntityType
}

type ParentAccountType struct {
	ID   int
	Name string `gorm:"not null;unique"`
}

type AccountType struct {
	ID                  int
	Name                string `gorm:"not null"`
	Description         string
	ParentAccountTypeID int `gorm:"not null"`
	ParentAccountType   ParentAccountType
}

type Account struct {
	gorm.Model
	Name          string `gorm:"not null;uniqueIndex:idx_name_account_type_entity"`
	Description   string
	AccountTypeID int `gorm:"not null;uniqueIndex:idx_name_account_type_entity"`
	AccountType   AccountType
	EntityID      int `gorm:"uniqueIndex:idx_name_account_type_entity"`
	Entity        Entity
}

type FinanceInstitutionAccount struct {
	gorm.Model
	Number               string
	AccountID            int
	Account              Account
	FinanceInstitutionID int
	FinanceInstitution   FinanceInstitution
}

type FinanceInstitutionType struct {
	gorm.Model
	Name string `gorm:"not null;unique"`
}

type FinanceInstitution struct {
	gorm.Model
	Name                     string `gorm:"not null;unique"`
	Description              string
	Number                   string
	FinanceInstitutionTypeID int
	FinanceInstitutionType   FinanceInstitutionType
}
