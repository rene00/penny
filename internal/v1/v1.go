// Package v1 provides code that is relevant to working with penny in it's original form, a webapp written in flask.
package v1

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"penny/internal/handler"
	"penny/internal/model"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"gorm.io/gorm"
)

// Config is the config struct for v1. It provides access to the flask config file.
type Config struct {
	SQLAlchemyDatabaseURI string
}

type Transaction struct {
	ID                int                `json:"id"`
	UserID            int                `json:"user_id"`
	Date              time.Time          `json:"date"`
	Debit             float64            `json:"debit"`
	Credit            float64            `json:"credit"`
	Memo              string             `json:"memo"`
	BankAccountID     int                `json:"bankaccount_id"`
	AccountID         int                `json:"account_id"`
	SplitTransactions []SplitTransaction `json:"split_transactions"`
}

type SplitTransaction struct {
	Transaction
	ParentID int `json:"parent_id"`
}

// NewConfig accepts a v1 config file and returns a Config.
func NewConfig(configFile string) (Config, error) {
	config := Config{}

	c, err := os.Open(configFile)
	if err != nil {
		return config, err
	}
	defer c.Close()

	scanner := bufio.NewScanner(c)
	var fileLines []string
	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Text())
	}

	for _, line := range fileLines {
		re := regexp.MustCompile(`^SQLALCHEMY_DATABASE_URI\s+=\s+(.*)$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			config.SQLAlchemyDatabaseURI = matches[1]
		}
	}

	if config.SQLAlchemyDatabaseURI == "" {
		return config, fmt.Errorf("unable to find SQLALCHEMY_DATABASE_URI in config file")
	}

	return config, nil
}

// trimQuotes accepts a string and returns the string with surrounding double
// and single quotes removed.
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// TransformSQLAlchemyDatabaseURI takes the SQLALCHEMY_DATABASE_URI string and
// formats to a DSN that works with go-sql-driver.
func TransformSQLAlchemyDatabaseURI(s string) (string, error) {
	re := regexp.MustCompile(`^{0,}mysql\+pymysql:\/\/(.*)\:(.*)\@(.*):(\d+)\/(.*)$`)
	m := re.FindStringSubmatch(trimQuotes(s))
	if len(m) != 6 {
		return "", fmt.Errorf("failed to parse URI")
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", m[1], m[2], m[3], m[4], m[5]), nil
}

// Validate checks the v1 DB to see if it can be migrated.
func Validate(ctx context.Context, v1db *sql.DB) error {
	if err := v1db.Ping(); err != nil {
		return err
	}

	// Check that only 1 user exists.
	var userCount int
	if err := v1db.QueryRowContext(ctx, "SELECT COUNT(1) FROM user").Scan(&userCount); err != nil {
		return err
	}
	if userCount != 1 {
		return fmt.Errorf("did not find exactly 1 user in v1 DB")
	}

	return nil
}

// migrateAccountTypes will migrate v1 parent accounts as ParentAccountType and
// v1 child accounts as AccountType. The difference between v1 and v2 is that
// parent accounts had parent_id set to NULL where as child accounts had a
// parent_id set.
func migrateAccountTypes(ctx context.Context, v1db *sql.DB, db *gorm.DB) error {
	dbHandler := handler.NewDBHandler()

	rows, err := v1db.QueryContext(ctx, "SELECT id, name FROM accounttype WHERE parent_id IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var parentID int
		if err = rows.Scan(&parentID, &name); err != nil {
			return err
		}

		parentAccountType := model.ParentAccountType{
			Name: name,
		}

		tx := db.WithContext(ctx)

		if err := tx.Create(&parentAccountType).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}

		rows2, err := v1db.QueryContext(ctx, "SELECT name FROM accounttype WHERE parent_id = ?", parentID)
		if err != nil {
			return err
		}
		defer rows2.Close()

		for rows2.Next() {
			accountType := model.AccountType{
				ParentAccountType: parentAccountType,
			}

			if err = rows2.Scan(&accountType.Name); err != nil {
				return err
			}

			if err := tx.Create(&accountType).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
				return err
			}
		}
	}

	return nil
}

func migrateEntityTypes(ctx context.Context, v1db *sql.DB, db *gorm.DB) error {
	dbHandler := handler.NewDBHandler()
	rows, err := v1db.QueryContext(ctx, "SELECT name FROM entitytype")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		entityType := &model.EntityType{}
		if err = rows.Scan(&entityType.Name); err != nil {
			return err
		}
		if err := db.WithContext(ctx).Create(entityType).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

func migrateEntities(ctx context.Context, v1db *sql.DB, db *gorm.DB, user model.User) error {
	dbHandler := handler.NewDBHandler()
	rows, err := v1db.QueryContext(ctx, `
SELECT e.name, et.name
	FROM entity AS e
	JOIN entitytype AS et ON e.entitytype_id = et.id
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var entityTypeName string
		if err = rows.Scan(&name, &entityTypeName); err != nil {
			return err
		}

		tx := db.WithContext(ctx)
		var entityType model.EntityType
		if err := tx.Where("name = ?", entityTypeName).First(&entityType).Error; err != nil {
			return err
		}

		entity := &model.Entity{
			Name:       name,
			User:       user,
			EntityType: entityType,
		}
		if err := tx.Create(entity).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

func migrateAccounts(ctx context.Context, v1db *sql.DB, db *gorm.DB, user model.User) error {
	rows, err := v1db.QueryContext(ctx, `
SELECT a.name, a.desc, at.name, e.name
	FROM account AS a
	JOIN entity AS e ON a.entity_id = e.id
	JOIN accounttype AS at ON a.accounttype_id = at.id
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var accountTypeName string
		var entityName string
		account := &model.Account{}
		if err = rows.Scan(&account.Name, &account.Description, &accountTypeName, &entityName); err != nil {
			return err
		}

		var accountType model.AccountType
		db.Where("name = ?", accountTypeName).First(&accountType)
		account.AccountType = accountType

		var entity model.Entity
		db.Where("name = ?", entityName).First(&entity)
		account.Entity = entity

		db.Create(account)
	}

	return nil
}

// migrateBankAccounts takes a v1 bank account and migrates it as an Account. A
// new FinanceInstituionAccount is created and linked to the Account.
func migrateBankAccounts(ctx context.Context, v1db *sql.DB, db *gorm.DB, user model.User) error {
	dbHandler := handler.NewDBHandler()

	rows, err := v1db.QueryContext(ctx, `
SELECT b.number, b.bank, b.desc, e.name
	FROM bankaccount AS b
	JOIN entity AS e ON b.entity_id = e.id
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {

		var bankAccountNumber string
		var bankAccountBank string
		var bankAccountDescription string
		entity := model.Entity{}

		if err = rows.Scan(
			&bankAccountNumber,
			&bankAccountBank,
			&bankAccountDescription,
			&entity.Name,
		); err != nil {
			return err
		}

		switch strings.ToLower(bankAccountBank) {
		case "personal cash", "olinda guest house - cash":
			// Skip cash accounts. When migrating transactions which belong to
			// cash accounts, they will be assigned to the default Cash account.
			continue
		}

		// Create the Account.
		account := model.Account{Entity: entity}
		account.Name = fmt.Sprintf("%s %s", bankAccountBank, bankAccountNumber)
		account.Description = bankAccountDescription

		tx := db.WithContext(ctx)

		if err := tx.Create(&account).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}

		// Create the FinanceInstitutionAccount and link it to Account.
		financeInstitutionAccount := model.FinanceInstitutionAccount{Number: bankAccountNumber, Account: account}

		financeIntitutionName := ""
		switch strings.ToLower(bankAccountBank) {
		case "cba", "mastercard platinum":
			financeIntitutionName = "Commonwealth Bank of Australia"
		case "paypal", "paypal ebb":
			financeIntitutionName = "PayPal"
		case "myer visa":
			financeIntitutionName = "Macquarie Bank"
		default:
			return fmt.Errorf("finance institution not supported:" + bankAccountBank)
		}

		var financeInstitution model.FinanceInstitution
		if err := tx.Where("name = ?", financeIntitutionName).First(&financeInstitution).Error; err != nil {
			return err
		}
		financeInstitutionAccount.FinanceInstitution = financeInstitution

		if err := tx.Create(&financeInstitutionAccount).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
			return err
		}
	}

	return nil
}

// Migrate migrates from the old v1 penny to the v2 penny db.
func Migrate(ctx context.Context, db *gorm.DB, configFile string) error {
	config, err := NewConfig(configFile)
	if err != nil {
		return err
	}

	uri, err := TransformSQLAlchemyDatabaseURI(config.SQLAlchemyDatabaseURI)
	if err != nil {
		return err
	}

	v1db, err := sql.Open("mysql", uri)
	if err != nil {
		return err
	}
	defer v1db.Close()

	if err := Validate(ctx, v1db); err != nil {
		return err
	}

	var user model.User
	db.Find(&user, 1)

	if err := migrateEntityTypes(ctx, v1db, db); err != nil {
		return err
	}

	if err := migrateEntities(ctx, v1db, db, user); err != nil {
		return err
	}

	if err := migrateAccountTypes(ctx, v1db, db); err != nil {
		return err
	}

	if err := migrateAccounts(ctx, v1db, db, user); err != nil {
		return err
	}

	if err := migrateBankAccounts(ctx, v1db, db, user); err != nil {
		return err
	}

	return nil
}
