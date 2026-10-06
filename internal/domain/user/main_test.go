package user_test

import (
	"os"
	"testing"

	"goilerplate/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

// TestMain hashes passwords at bcrypt's minimum cost. At the production cost one hash takes
// about two seconds under -race, and these tests hash on almost every case. What they check is
// the flow, not bcrypt's work factor; pkg/utils tests that the cost is applied.
func TestMain(m *testing.M) {
	utils.SetPasswordCostForTests(bcrypt.MinCost)
	os.Exit(m.Run())
}
