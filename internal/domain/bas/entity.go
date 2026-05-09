package bas

import (
	"strings"

	"goilerplate/pkg/utils"
)

type Bas struct {
	ID   string
	Code string
	Name string
}

func (e *Bas) validate() error {
	code := strings.ToUpper(strings.TrimSpace(e.Code))
	if code == "" {
		return utils.ClientErr(400, "code is required")
	}
	if strings.TrimSpace(e.Name) == "" {
		return utils.ClientErr(400, "name is required")
	}
	return nil
}

func (e *Bas) Clone() *Bas {
	return &Bas{
		ID:   e.ID,
		Code: e.Code,
		Name: e.Name,
	}
}
