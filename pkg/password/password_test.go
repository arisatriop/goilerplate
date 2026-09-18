package password_test

import (
	"net/http"
	"strings"
	"testing"

	"goilerplate/pkg/password"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listChecker stands in for a deployment-supplied common-password list.
type listChecker map[string]bool

func (c listChecker) IsCommon(p string) bool { return c[p] }

func assertRejected(t *testing.T, err error, contains string) {
	t.Helper()

	var clientErr *utils.ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Equal(t, http.StatusBadRequest, clientErr.Code, "choosing a password is request input")
	assert.Contains(t, clientErr.Error(), contains)
}

func TestPolicy_Length(t *testing.T) {
	policy := password.NewPolicy(nil)

	assertRejected(t, policy.Validate("short"), "at least 8 characters")
	assert.NoError(t, policy.Validate("12345678"), "exactly the minimum is allowed")

	// bcrypt truncates past 72 bytes, so a longer password must be refused rather than
	// silently shortened.
	assert.NoError(t, policy.Validate(strings.Repeat("a", password.MaxBytes)))
	assertRejected(t, policy.Validate(strings.Repeat("a", password.MaxBytes+1)), "at most 72 bytes")

	// The cap is bytes, not characters: 40 three-byte runes are 120 bytes.
	assertRejected(t, policy.Validate(strings.Repeat("日", 40)), "at most 72 bytes")
	assert.NoError(t, policy.Validate(strings.Repeat("日", 24)), "24 runes is 72 bytes")
}

// No composition rules: NIST 800-63B drops them because they push people toward predictable
// substitutions without adding real entropy.
func TestPolicy_NoCompositionRules(t *testing.T) {
	policy := password.NewPolicy(nil)

	for _, p := range []string{"alllowercase", "ALLUPPERCASE", "1234567890", "          "} {
		assert.NoError(t, policy.Validate(p), "composition must not be enforced: %q", p)
	}
}

func TestPolicy_CommonPasswordsRejected(t *testing.T) {
	policy := password.NewPolicy(listChecker{"password123": true})

	assertRejected(t, policy.Validate("password123"), "too common")
	assert.NoError(t, policy.Validate("a-less-predictable-one"))
}

// The default accepts everything, so the policy works out of the box; a deployment that cares
// supplies a list.
func TestPolicy_DefaultHasNoList(t *testing.T) {
	assert.NoError(t, password.NewPolicy(nil).Validate("password123"))
	assert.False(t, password.NoCommonList{}.IsCommon("password123"))
}
