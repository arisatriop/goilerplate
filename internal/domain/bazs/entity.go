package bazs

import (
	"strings"

	"goilerplate/pkg/utils"
)

type Bazs struct {
	ID   string
	Code string
	Name string
}

func (e *Bazs) validate() error {
	code := strings.ToUpper(strings.TrimSpace(e.Code))
	if code == "" {
		return utils.ClientErr(400, "code is required")
	}
	if strings.TrimSpace(e.Name) == "" {
		return utils.ClientErr(400, "name is required")
	}
	return nil
}

func (e *Bazs) Clone() *Bazs {
	return &Bazs{
		ID:   e.ID,
		Code: e.Code,
		Name: e.Name,
	}
}
