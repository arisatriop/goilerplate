package zexample

import (
	"goilerplate/pkg/utils"
)

type Example struct {
	ID      string
	Code    string
	Example string
}

func (e *Example) validate() error {

	firstThree := e.Code[:3]
	if firstThree != "EXP" {
		return utils.Error(400, "code must start with 'EXP'")
	}

	return nil
}

func (e *Example) Clone() *Example {
	return &Example{
		ID:      e.ID,
		Code:    e.Code,
		Example: e.Example,
	}
}
