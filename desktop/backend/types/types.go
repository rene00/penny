package types

type JSResp struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    any    `json:"data,omitempty"`
}

type Currency struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
