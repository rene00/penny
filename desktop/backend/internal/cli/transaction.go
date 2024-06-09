package cli

import (
	"fmt"
	"os"
	"pennyapp/backend/config"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/services"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func transactionCmd(cli *cli) *cobra.Command {
	var cmd = &cobra.Command{
		Use: "transaction",
	}
	cmd.AddCommand(transactionUpdateCmd(cli))
	return cmd
}

func transactionUpdateCmd(cli *cli) *cobra.Command {
	var flags struct {
		transactionID int64
		memo          string
		date          string
	}
	var cmd = &cobra.Command{
		Use: "update",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("transaction-id", cmd.Flags().Lookup("transaction-id"))
			_ = viper.BindPFlag("memo", cmd.Flags().Lookup("memo"))
			_ = viper.BindPFlag("date", cmd.Flags().Lookup("date"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logwrap.New("logger", os.Stdout, true)
			logger.SetLevel(logwrap.INFO)

			configOpts := []config.OptFunc{config.WithDebug(cli.debug)}
			if cli.debug {
				logger.SetLevel(logwrap.DEBUG)
				boil.DebugMode = true
			}

			ctx := cmd.Context()
			conf := config.NewConfig(configOpts...)
			transactionSvc := services.Transaction()
			dbSvc := services.DB()
			dbSvc.Start(ctx, conf)
			transactionSvc.Start(ctx, dbSvc.DB, logger)

			transaction, err := model.FindTransaction(ctx, dbSvc.DB, flags.transactionID)
			if err != nil {
				return err
			}

			memo := transaction.Memo
			if flags.memo != "" {
				memo = flags.memo
			}

			date, err := services.GetVUSJSdateFromSQLiteDate(transaction.Date)
			if err != nil {
				return err
			}

			if flags.date != "" {
				date = flags.date
			}

			resp := transactionSvc.UpdateTransaction(transaction.ID, memo, date)
			fmt.Printf("resp: %#+v\n", resp)

			return nil
		},
	}

	cmd.Flags().Int64Var(&flags.transactionID, "transaction-id", 0, "transaction id")
	cmd.Flags().StringVar(&flags.memo, "memo", "", "memo")
	cmd.Flags().StringVar(&flags.date, "date", "", "date")

	return cmd
}
