package types

type JSResp struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    any    `json:"data,omitempty"`
}

type Account struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	AccountType AccountType `json:"account_type"`
	Amount      string      `json:"amount"`
}

type AccountType struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	*ParentAccountType `json:"parent_account_type"`
}

type ParentAccountType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Currency struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
