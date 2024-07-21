package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"pennyapp/backend/config"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/services"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type cli struct {
	debug bool
}

func Execute() {
	cli := &cli{}
	rootCmd := buildRootCmd(cli)
	rootCmd.AddCommand(importTransactionsCmd(cli))
	if err := rootCmd.ExecuteContext(context.TODO()); err != nil {
		os.Exit(1)
	}
}

func buildRootCmd(cli *cli) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "dt",
	}
	rootCmd.AddCommand(accountCmd(cli))
	rootCmd.AddCommand(transactionCmd(cli))
	rootCmd.PersistentFlags().BoolVar(&cli.debug, "debug", false, "debug")
	return rootCmd
}

func importTransactionsCmd(cli *cli) *cobra.Command {
	var flags struct {
		ofxFilename string
	}

	var cmd = &cobra.Command{
		Use: "import-transactions",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("ofx-filename", cmd.Flags().Lookup("ofx-filename"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			logger := logwrap.New("logger", os.Stdout, true)
			logger.SetLevel(logwrap.INFO)

			if cli.debug {
				logger.SetLevel(logwrap.DEBUG)
				boil.DebugMode = true
			}

			b, err := os.ReadFile(flags.ofxFilename)
			if err != nil {
				return err
			}

			enc := base64.StdEncoding.EncodeToString(b)

			// javascript prepends `data:application/octet-stream;base64,`
			jsPrepend := "data:application/octet-stream;base64,"
			s := fmt.Sprintf("%s%s", jsPrepend, enc)

			ctx := cmd.Context()

			transactionSvc := services.Transaction()
			transactionImporterSvc := services.TransactionImporter()
			accountSvc := services.Account()
			dbSvc := services.DB()

			conf, err := config.NewConfig(config.WithDebug(cli.debug))
			if err != nil {
				return err
			}

			dbSvc.Start(ctx, conf)

			defaultResources, err := defaultresources.NewDefaultResources(ctx, dbSvc.DB)
			if err != nil {
				return err
			}

			accountSvc.Start(ctx, dbSvc.DB, logger, defaultResources)

			transactionImporterSvc.Start(ctx, conf, dbSvc.DB, logger, defaultResources)

			transactionSvc.Start(ctx, dbSvc.DB, logger, transactionImporterSvc)
			transactionSvc.AddAccountSvcHandler(accountSvc)

			resp := transactionSvc.ImportTransactions(s)
			fmt.Printf("resp: %#+v\n", resp)

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.ofxFilename, "ofx-filename", "", "ofx filename")

	return cmd
}
