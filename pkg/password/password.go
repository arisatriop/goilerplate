// Package password holds the password rules this project applies wherever a user chooses a
// password: registration, change, and reset. Keeping them in one place means the three paths
// cannot drift apart.
//
// The rules follow NIST SP 800-63B: length is what matters, and composition rules
// (an uppercase, a digit, a symbol) are deliberately absent because they push people toward
// predictable substitutions without adding real entropy.
package password

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"goilerplate/pkg/utils"
)

const (
	// MinLength is the NIST 800-63B minimum for a user-chosen password.
	MinLength = 8

	// MaxBytes is bcrypt's hard limit. Anything longer is silently truncated by the
	// algorithm, so a 100-character password would be no stronger than its first 72 bytes
	// while giving the user the opposite impression. Counted in bytes, not characters:
	// a multi-byte character costs more than one.
	MaxBytes = 72
)

// CommonChecker reports whether a password appears in a list of commonly used or compromised
// passwords, which NIST 800-63B requires checking against. Implementations back this with
// whatever list a deployment chooses; none ships with the boilerplate.
type CommonChecker interface {
	IsCommon(password string) bool
}

// NoCommonList is the default: it accepts every password. It exists so the policy works out
// of the box, not because skipping the check is a good idea. A deployment that cares should
// supply a real list.
type NoCommonList struct{}

// IsCommon always reports false.
func (NoCommonList) IsCommon(string) bool { return false }

// Policy validates a chosen password.
type Policy struct {
	common CommonChecker
}

// NewPolicy builds a Policy. A nil checker falls back to NoCommonList.
func NewPolicy(common CommonChecker) Policy {
	if common == nil {
		common = NoCommonList{}
	}
	return Policy{common: common}
}

// Validate reports why a password is unacceptable, as a 400 client error. Choosing a password
// is request input, so a rejection is a validation error rather than an authentication one.
func (p Policy) Validate(password string) error {
	if count := utf8.RuneCountInString(password); count < MinLength {
		return utils.ClientErr(http.StatusBadRequest,
			fmt.Sprintf("Password must be at least %d characters", MinLength))
	}

	// Reported before hashing, because bcrypt would otherwise truncate silently and the user
	// would never learn that most of what they typed was ignored.
	if len(password) > MaxBytes {
		return utils.ClientErr(http.StatusBadRequest,
			fmt.Sprintf("Password must be at most %d bytes", MaxBytes))
	}

	if p.common.IsCommon(password) {
		return utils.ClientErr(http.StatusBadRequest,
			"Password is too common, please choose a less predictable one")
	}

	return nil
}
