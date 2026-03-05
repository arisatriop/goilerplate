package foo

import (
	"goilerplate/pkg/utils"
)

type Foo struct {
	ID       string
	Code     string
	Foo string
}

func (e *Foo) validate() error {

	firstThree := e.Code[:3]
	if firstThree != "TMP" {
		return utils.ClientErr(400, "Code must start with 'TMP'")
	}

	return nil
}

func (e *Foo) Clone() *Foo {
	return &Foo{
		ID:       e.ID,
		Code:     e.Code,
		Foo: e.Foo,
	}
}
