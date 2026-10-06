package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// withPasswordCost sets the cost for one test and restores the previous one afterwards. Tests
// using it must not run in parallel: the cost is process-wide.
func withPasswordCost(t *testing.T, cost int) {
	t.Helper()
	previous := int(passwordCost.Load())
	SetPasswordCostForTests(cost)
	t.Cleanup(func() { SetPasswordCostForTests(previous) })
}

// The regression: the dummy hash was a constant at cost 10 while passwords were hashed at 12, so
// an unknown email was answered about four times faster than a wrong password, and response
// time alone said which addresses were registered. Checked at the production cost, since that
// is the pairing that leaked.
func TestSimulatePasswordCheck_CostsTheSameAsARealHash(t *testing.T) {
	withPasswordCost(t, DefaultCost)

	real, err := HashPassword("a-real-password")
	require.NoError(t, err)

	realCost, err := bcrypt.Cost([]byte(real))
	require.NoError(t, err)
	dummyCost, err := bcrypt.Cost(dummyHash())
	require.NoError(t, err)

	assert.Equal(t, DefaultCost, realCost)
	assert.Equal(t, realCost, dummyCost, "a cheaper dummy makes unknown emails answer faster")
}

// The dummy follows a change of cost, so it cannot drift from HashPassword in either direction.
func TestSimulatePasswordCheck_FollowsTheCost(t *testing.T) {
	withPasswordCost(t, bcrypt.MinCost)

	real, err := HashPassword("a-real-password")
	require.NoError(t, err)
	realCost, err := bcrypt.Cost([]byte(real))
	require.NoError(t, err)
	dummyCost, err := bcrypt.Cost(dummyHash())
	require.NoError(t, err)

	assert.Equal(t, bcrypt.MinCost, realCost)
	assert.Equal(t, bcrypt.MinCost, dummyCost)
}

func TestSimulatePasswordCheck_MatchesNothing(t *testing.T) {
	withPasswordCost(t, bcrypt.MinCost)

	for _, password := range []string{"", "password", "a-real-password"} {
		assert.Error(t, bcrypt.CompareHashAndPassword(dummyHash(), []byte(password)))
	}
	SimulatePasswordCheck("anything") // must not panic
}
