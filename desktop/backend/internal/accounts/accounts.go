package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"pennyapp/backend/defaultresources"
	"pennyapp/backend/internal/hashers"
	"pennyapp/backend/model"
	"regexp"

	"github.com/aclindsa/ofxgo"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func getSurpressedCreditCardNumber(s string) (string, error) {
	re := regexp.MustCompile(`^\d{16}$`)
	if !re.MatchString(s) {
		return "", fmt.Errorf("not a credit card")
	}

	rs := []rune(s)
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = 'X'
	}

	return string(rs), nil
}

func CreateCreditCardAccount(ctx context.Context, db *sql.DB, defaultResources defaultresources.DefaultResources, cc ofxgo.CCAcct) (model.Account, error) {
	var account model.Account
	var err error

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return account, err
	}

	accountType, err := model.AccountTypes(qm.Where("name=?", "Liability")).One(ctx, tx)
	if err != nil {
		return account, err
	}

	ccSurpressed, err := getSurpressedCreditCardNumber(cc.AcctID.String())
	if err != nil {
		return account, err
	}

	account = model.Account{
		Name:          fmt.Sprintf("Credit Card %s", ccSurpressed),
		AccountTypeID: accountType.ID,
		EntityID:      defaultResources.EntityID(),
		ParentID:      defaultResources.GrandParentAccountID(),
	}

	if err := account.Insert(ctx, tx, boil.Infer()); err != nil {
		return account, err
	}

	accountAttributes := []model.AccountAttribute{
		{
			AccountID: account.ID,
			Name:      "cc_account_id_hash",
			Value:     getCreditCardHash(cc),
		},
	}

	for _, i := range accountAttributes {
		if err := i.Insert(ctx, tx, boil.Infer()); err != nil {
			return account, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return account, err
	}

	return account, nil
}

func getCreditCardHash(cc ofxgo.CCAcct) string {
	return fmt.Sprintf("%x", hashers.NewCreditCardHash().Hash(cc.AcctID.String()).Sum(nil))
}
