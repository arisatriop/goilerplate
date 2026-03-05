package bar

import (
	"goilerplate/pkg/utils"
)

type Bar struct {
	ID      string
	Code    string
	Bar string
}

func (e *Bar) validate() error {

	firstThree := e.Code[:3]
	if firstThree != "EXP" {
		return utils.ClientErr(400, "code must start with 'EXP'")
	}

	return nil
}

func (e *Bar) Clone() *Bar {
	return &Bar{
		ID:      e.ID,
		Code:    e.Code,
		Bar: e.Bar,
	}
}
