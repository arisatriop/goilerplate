package zexamplenew

import "goilerplate/pkg/utils"

type ZexampleNew struct {
	ID      string
	Code    string
	Example string
}

func (e *ZexampleNew) validate() error {

	firstThree := e.Code[:3]
	if firstThree != "EXP" {
		return utils.Error(400, "code must start with 'EXP'")
	}

	return nil
}

func (e *ZexampleNew) Clone() *ZexampleNew {
	return &ZexampleNew{
		ID:      e.ID,
		Code:    e.Code,
		Example: e.Example,
	}
}
