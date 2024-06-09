package types

import (
	"fmt"
	"pennyapp/backend/model"
)

type SplitResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    Split  `json:"data"`
}

func NewSplitResponse(s model.Split) SplitResponse {
	split := Split{
		ID:         s.ID,
		ValueNum:   s.ValueNum,
		ValueDenom: s.ValueDenom,
		Amount:     getAmountAsString(s.ValueNum, s.ValueDenom),
		Account: Account{
			ID:   s.R.Account.ID,
			Name: s.R.Account.Name,
			AccountType: AccountType{
				ID:   s.R.Account.R.AccountType.ID,
				Name: s.R.Account.R.AccountType.Name,
			},
		},
	}

	r := SplitResponse{
		Success: false,
		Msg:     "default",
		Data:    split,
	}
	return r
}

type Split struct {
	ID         int64   `json:"id"`
	ValueNum   int64   `json:"value_num"`
	ValueDenom int64   `json:"value_denom"`
	Amount     string  `json:"amount"`
	Account    Account `json:"account"`
}

func NewSplit(s *model.Split) (Split, error) {
	var split Split

	if s.R.Account == nil {
		return split, fmt.Errorf("missing split account relationship")
	}

	split = Split{
		ID:         s.ID,
		ValueNum:   s.ValueNum,
		ValueDenom: s.ValueDenom,
	}

	split.Account = Account{
		ID:   s.R.Account.ID,
		Name: s.R.Account.Name,
	}
	return split, nil
}
