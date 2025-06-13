package model

type Params struct {
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

func DefaultParams() Params {
	return Params{
		Limit:  10,
		Offset: 0,
	}
}
