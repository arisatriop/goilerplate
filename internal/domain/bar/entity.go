package bar

import (
	"strings"
)

type Bar struct {
	ID   string
	Code string
	Bar  string
}

func (e *Bar) validate() error {
	code := normalizeCode(e.Code)
	if code == "" {
		return ErrCodeRequired
	}
	if len(code) < 3 || !strings.HasPrefix(code, "EXP") {
		return ErrCodeFormat
	}
	if strings.TrimSpace(e.Bar) == "" {
		return ErrDescriptionRequired
	}
	return nil
}

// normalize puts the fields in their stored form. Code comparisons, including the database's
// uniqueness check, happen on this form, so "exp-1" and " EXP-1 " are the same code.
func (e *Bar) normalize() {
	e.Code = normalizeCode(e.Code)
	e.Bar = strings.TrimSpace(e.Bar)
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func (e *Bar) Clone() *Bar {
	return &Bar{
		ID:   e.ID,
		Code: e.Code,
		Bar:  e.Bar,
	}
}
