package cli

import (
	"fmt"
	"os"
	"pennyapp/backend/config"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/services"
	"pennyapp/backend/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func accountCmd(cli *cli) *cobra.Command {
	var cmd = &cobra.Command{
		Use: "account",
	}
	cmd.AddCommand(accountListCmd(cli))
	cmd.AddCommand(accountUpdateCmd(cli))
	return cmd
}

func accountUpdateCmd(cli *cli) *cobra.Command {
	var flags struct {
		accountID     int64
		description   string
		name          string
		accountTypeID int64
	}
	var cmd = &cobra.Command{
		Use: "update",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("account-id", cmd.Flags().Lookup("account-id"))
			_ = viper.BindPFlag("description", cmd.Flags().Lookup("description"))
			_ = viper.BindPFlag("name", cmd.Flags().Lookup("name"))
			_ = viper.BindPFlag("account-type-id", cmd.Flags().Lookup("account-type-id"))
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

			conf, err := config.NewConfig(configOpts...)
			if err != nil {
				return err
			}

			accountSvc := services.Account()
			dbSvc := services.DB()
			dbSvc.Start(ctx, conf)

			defaultResources, err := defaultresources.NewDefaultResources(ctx, dbSvc.DB)
			if err != nil {
				return err
			}

			accountSvc.Start(ctx, dbSvc.DB, logger, defaultResources)

			account, err := model.FindAccount(ctx, dbSvc.DB, flags.accountID)
			if err != nil {
				return err
			}

			if flags.accountTypeID != 0 {
				account.AccountTypeID = flags.accountTypeID
			}

			resp := accountSvc.UpdateAccount(account.ID, flags.name, flags.description, account.AccountTypeID)
			fmt.Printf("resp: %#+v\n", resp)

			return nil
		},
	}

	cmd.Flags().Int64Var(&flags.accountID, "account-id", 0, "account id")
	cmd.Flags().StringVar(&flags.description, "description", "", "description")
	cmd.Flags().StringVar(&flags.name, "name", "", "name")
	cmd.Flags().Int64Var(&flags.accountTypeID, "account-type-id", 0, "account type id")
	return cmd
}

func accountListCmd(cli *cli) *cobra.Command {
	var cmd = &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logwrap.New("logger", os.Stdout, true)
			logger.SetLevel(logwrap.INFO)

			configOpts := []config.OptFunc{config.WithDebug(cli.debug)}
			if cli.debug {
				logger.SetLevel(logwrap.DEBUG)
				boil.DebugMode = true
			}

			ctx := cmd.Context()

			accountSvc := services.Account()
			dbSvc := services.DB()

			conf, err := config.NewConfig(configOpts...)
			if err != nil {
				return err
			}

			dbSvc.Start(ctx, conf)

			defaultResources, err := defaultresources.NewDefaultResources(ctx, dbSvc.DB)
			if err != nil {
				return err
			}

			accountSvc.Start(ctx, dbSvc.DB, logger, defaultResources)

			resp := accountSvc.GetAccounts()
			fmt.Printf("Success: %t\n", resp.Success)
			fmt.Printf("Msg: %s", resp.Msg)
			switch data := resp.Data.(type) {
			case []types.Account:
				for _, account := range data {
					fmt.Printf("Account\n")
					fmt.Printf("ID: %#+v\n", account.ID)
				}
			}
			return nil
		},
	}
	return cmd
}
