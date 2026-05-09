package bas

import (
	"strings"

	"goilerplate/pkg/utils"
)

type Bas struct {
	ID   string
	Code string
	Bas  string
}

func (e *Bas) validate() error {
	if strings.TrimSpace(e.Code) == "" {
		return utils.ClientErr(400, "code is required")
	}
	if strings.TrimSpace(e.Bas) == "" {
		return utils.ClientErr(400, "bas is required")
	}
	return nil
}

func (e *Bas) Clone() *Bas {
	return &Bas{
		ID:   e.ID,
		Code: e.Code,
		Bas:  e.Bas,
	}
}
