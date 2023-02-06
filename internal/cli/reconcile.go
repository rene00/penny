package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	v1 "penny/internal/v1"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ExecuteReconcile() {
	var flags struct {
		DryRun              bool
		ReconcileFile       string
		ParentTransactionID int
	}

	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "reconcile a transaction",
		Long: `
$ reconcile --reconcile-file f.json --parent-transaction-id 42
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("reconcile-file", cmd.Flags().Lookup("reconcile-file"))
			_ = viper.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))
			_ = viper.BindPFlag("parent-transaction-id", cmd.Flags().Lookup("parent-transaction-id"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			configFile, ok := os.LookupEnv("CONFIG_FILE")
			if !ok {
				return fmt.Errorf("CONFIG_FILE not set")
			}

			config, err := v1.NewConfig(configFile)
			if err != nil {
				return err
			}

			uri, err := v1.TransformSQLAlchemyDatabaseURI(config.SQLAlchemyDatabaseURI)
			if err != nil {
				return err
			}

			db, err := sql.Open("mysql", uri)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := v1.Validate(ctx, db); err != nil {
				return err
			}

			f, err := os.Open(flags.ReconcileFile)
			if err != nil {
				return err
			}
			defer f.Close()

			b, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			var rt v1.Transaction
			if err := json.Unmarshal(b, &rt); err != nil {
				return err
			}

			// Find transaction from parent-transaction-id
			var t v1.Transaction
			if err := db.QueryRowContext(ctx, "SELECT id, user_id, date, debit, credit, memo, bankaccount_id, account_id FROM tx WHERE id = ?", flags.ParentTransactionID).Scan(&t.ID, &t.UserID, &t.Date, &t.Debit, &t.Credit, &t.Memo, &t.BankAccountID, &t.AccountID); err != nil {
				return fmt.Errorf("failed to select transaction: %w", err)
			}

			if err := v1.ReconcileTransaction(ctx, db, rt, t); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", true, "Dry run")
	cmd.Flags().StringVar(&flags.ReconcileFile, "reconcile-file", "", "Reconcile file")
	cmd.Flags().IntVar(&flags.ParentTransactionID, "parent-transaction-id", 0, "Parent transaction ID")

	cmd.MarkFlagRequired("reconcile-file")
	cmd.MarkFlagRequired("parent-transaction-id")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
