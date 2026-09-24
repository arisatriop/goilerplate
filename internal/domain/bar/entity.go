package bar

import (
	"strings"

	"goilerplate/pkg/utils"
)

type Bar struct {
	ID   string
	Code string
	Bar  string
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

// normalize puts the fields in their stored form. Code comparisons, including the database's
// uniqueness check, happen on this form, so "exp-1" and " EXP-1 " are the same code.
func (e *Bar) normalize() {
	e.Code = strings.ToUpper(strings.TrimSpace(e.Code))
	e.Bar = strings.TrimSpace(e.Bar)
}

func (e *Bar) Clone() *Bar {
	return &Bar{
		ID:   e.ID,
		Code: e.Code,
		Bar:  e.Bar,
	}
}
