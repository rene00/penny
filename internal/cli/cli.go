// Package cli is a package.
package cli

import (
	"context"
	"fmt"
	"os"
	"path"
	"penny/internal/handler"
	"penny/internal/model"
	"penny/internal/seed"
	v1 "penny/internal/v1"
	"strings"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var homeDir = os.Getenv("HOME")
var defaultPennyDBFile = path.Join(homeDir, ".config/penny.db")

// Config is the main configuration struct.
type Config struct {
	DBFile     string
	DB         *gorm.DB
	OldMigrate bool
}

type cli struct {
	initOnce sync.Once
	errOnce  error
	config   Config
}

func (c *cli) init(ctx context.Context) error {
	c.initOnce.Do(func() {
		if c.errOnce = c.initContext(ctx); c.errOnce != nil {
			return
		}
	})
	return c.errOnce
}

func (c *cli) initContext(ctx context.Context) error {
	c.config.DBFile = defaultPennyDBFile
	dbFile, ok := os.LookupEnv("PENNY_DB_FILE")
	if ok {
		c.config.DBFile = dbFile
	}

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "p_",
		},
	}

	db, err := gorm.Open(sqlite.Open(c.config.DBFile), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to db")
	}
	c.config.DB = db

	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.EntityType{})
	db.AutoMigrate(&model.Entity{})
	db.AutoMigrate(&model.ParentAccountType{})
	db.AutoMigrate(&model.AccountType{})
	db.AutoMigrate(&model.Account{})
	db.AutoMigrate(&model.FinanceInstitutionType{})
	db.AutoMigrate(&model.FinanceInstitution{})
	db.AutoMigrate(&model.FinanceInstitutionAccount{})

	dbHandler := handler.NewDBHandler()

	// Create the default user.
	if err := db.Create(&model.User{Name: "Default User"}).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
		return err
	}

	// Create the default entity.
	if err := db.Create(&model.Entity{Name: "Personal"}).Error; err != nil && !dbHandler.IsSQLiteErrConstraintUnique(err) {
		return err
	}

	seeder := seed.NewSeed(db)

	if err := seeder.FinanceInstitutionType(ctx); err != nil {
		return err
	}

	if err := seeder.FinanceInstitution(ctx); err != nil {
		return err
	}

	if err := seeder.ParentAccountType(ctx); err != nil {
		return err
	}

	if err := seeder.AccountType(ctx); err != nil {
		return err
	}

	if err := seeder.Account(ctx); err != nil {
		return err
	}

	migrateV2 := os.Getenv("PENNY_MIGRATE_V1")
	if strings.ToLower(migrateV2) == "true" {
		configFile, ok := os.LookupEnv("CONFIG_FILE")
		if !ok {
			return fmt.Errorf("CONFIG_FILE not set")
		}
		if err := v1.Migrate(ctx, c.config.DB, configFile); err != nil {
			return err
		}
	}

	return nil
}

// Execute is the main entry point.
func Execute() {
	cli := &cli{}
	ctx := context.Background()
	if err := cli.init(ctx); err != nil {
		panic(err)
	}
}
