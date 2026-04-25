package bar

import (
	"strings"

	"goilerplate/pkg/utils"
)

type Bar struct {
	ID      string
	Code    string
	Bar string
}

func (e *Bar) validate() error {
	code := strings.ToUpper(strings.TrimSpace(e.Code))
	if code == "" {
		return utils.ClientErr(400, "code is required")
	}
	if len(code) < 3 || !strings.HasPrefix(code, "EXP") {
		return utils.ClientErr(400, "code must start with 'EXP'")
	}
	if strings.TrimSpace(e.Bar) == "" {
		return utils.ClientErr(400, "bar is required")
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
