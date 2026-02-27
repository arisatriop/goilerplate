package template

import (
	"goilerplate/pkg/utils"
)

type Template struct {
	ID       string
	Code     string
	Template string
}

func (e *Template) validate() error {

	firstThree := e.Code[:3]
	if firstThree != "TMP" {
		return utils.Error(400, "Code must start with 'TMP'")
	}

	return nil
}

func (e *Template) Clone() *Template {
	return &Template{
		ID:       e.ID,
		Code:     e.Code,
		Template: e.Template,
	}
}
