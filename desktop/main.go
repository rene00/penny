package main

import (
	"context"
	"embed"
	"flag"
	"os"

	"pennyapp/backend/config"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	var err error

	logger := logwrap.New("logger", os.Stdout, true)
	logger.SetLevel(logwrap.INFO)

	var (
		debugFlag = flag.Bool("debug", false, "debug")
		dbFile    = flag.String("db-file", "", "sqlite3 database file")
	)

	flag.Parse()

	configOpts := []config.OptFunc{
		config.WithDBFile(*dbFile),
		config.WithDebug(*debugFlag),
	}

	if *debugFlag {
		logger.SetLevel(logwrap.DEBUG)
	}

	conf, err := config.NewConfig(configOpts...)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	dbSvc := services.DB()
	transactionImporterSvc := services.TransactionImporter()
	accountSvc := services.Account()
	entitySvc := services.Entity()
	transactionSvc := services.Transaction()
	splitsSvc := services.Splits()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "Penny",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			dbSvc.Start(ctx, conf)

			defaultResources, err := defaultresources.NewDefaultResources(ctx, dbSvc.DB)
			if err != nil {
				println("Failed to find default resources: ", err.Error())
				panic(err.Error())
			}

			transactionImporterSvc.Start(ctx, conf, dbSvc.DB, logger, defaultResources)

			entitySvc.Start(ctx, dbSvc.DB)
			accountSvc.Start(ctx, dbSvc.DB, logger, defaultResources)
			transactionSvc.Start(ctx, dbSvc.DB, logger, transactionImporterSvc)
			splitsSvc.Start(ctx, dbSvc.DB, logger)

			transactionSvc.AddAccountSvcHandler(accountSvc)

			app.startup(ctx)
		},
		Bind: []interface{}{
			dbSvc,
			entitySvc,
			accountSvc,
			transactionSvc,
			splitsSvc,
		},
	})

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
