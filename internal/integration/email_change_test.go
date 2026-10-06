package integration

import (
	"net/http"
	"testing"

	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type emailChangeRequest struct {
	NewEmail        string `json:"newEmail"`
	CurrentPassword string `json:"currentPassword"`
}

// The whole change against the real SQL: the code goes to the new inbox, the account moves only
// once it is confirmed, the old address is warned, and sign-in follows the address.
func TestEmailChange_MovesTheAccount(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	oldEmail := s.createUser(t, testPassword)
	newEmail := utils.GenerateUUID() + "@example.test"
	session := s.login(t, oldEmail, testPassword)

	status, body := s.do(t, http.MethodPost, "/email-change", session.Access, emailChangeRequest{newEmail, testPassword})
	require.Equal(t, http.StatusOK, status, "email-change: %s", body)
	code := s.inbox.latestChangeCode(newEmail)
	require.NotEmpty(t, code, "the code goes to the new address")

	s.login(t, oldEmail, testPassword) // nothing has changed yet

	status, body = s.do(t, http.MethodPost, "/email-change/confirm", session.Access, map[string]string{"code": code})
	require.Equal(t, http.StatusOK, status, "confirm: %s", body)

	status, _ = s.do(t, http.MethodPost, "/login", "", loginRequest{Email: oldEmail, Password: testPassword})
	assert.Equal(t, http.StatusUnauthorized, status, "the old address no longer signs in")
	s.login(t, newEmail, testPassword)

	status, _ = s.do(t, http.MethodGet, "/me", session.Access, nil)
	assert.Equal(t, http.StatusOK, status, "the session that made the change is kept")

	notices := s.inbox.changedNotices(oldEmail)
	require.Len(t, notices, 1, "the old inbox is warned")
	assert.Equal(t, newEmail, notices[0].NewEmail)

	var verified bool
	require.NoError(t, s.db.Raw("SELECT email_verified FROM users WHERE email = ?", newEmail).Scan(&verified).Error)
	assert.True(t, verified)
}

// A taken address is answered like a free one, so a signed-in user cannot probe the user list.
func TestEmailChange_TakenAddressSendsNothing(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	email := s.createUser(t, testPassword)
	taken := s.createUser(t, testPassword)
	session := s.login(t, email, testPassword)

	status, _ := s.do(t, http.MethodPost, "/email-change", session.Access, emailChangeRequest{taken, testPassword})

	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, s.inbox.latestChangeCode(taken))
}

// The address was free when the code was sent and registered before it was confirmed: the
// unique constraint refuses the move, and the account keeps its address.
func TestEmailChange_AddressTakenBeforeConfirming(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	email := s.createUser(t, testPassword)
	newEmail := utils.GenerateUUID() + "@example.test"
	session := s.login(t, email, testPassword)

	status, _ := s.do(t, http.MethodPost, "/email-change", session.Access, emailChangeRequest{newEmail, testPassword})
	require.Equal(t, http.StatusOK, status)
	code := s.inbox.latestChangeCode(newEmail)

	other := s.createUser(t, testPassword)
	require.NoError(t, s.db.Exec("UPDATE users SET email = ? WHERE email = ?", newEmail, other).Error)

	status, body := s.do(t, http.MethodPost, "/email-change/confirm", session.Access, map[string]string{"code": code})
	assert.Equal(t, http.StatusConflict, status)
	assertCode(t, body, "email_already_registered")
	s.login(t, email, testPassword)
}

func TestEmailChange_RequiresTheCurrentPassword(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	email := s.createUser(t, testPassword)
	session := s.login(t, email, testPassword)
	newEmail := utils.GenerateUUID() + "@example.test"

	status, body := s.do(t, http.MethodPost, "/email-change", session.Access, emailChangeRequest{newEmail, "not-the-password"})

	assert.Equal(t, http.StatusUnauthorized, status)
	assertCode(t, body, "invalid_credentials")
	assert.Empty(t, s.inbox.latestChangeCode(newEmail))

	status, _ = s.do(t, http.MethodPost, "/email-change", "", emailChangeRequest{newEmail, testPassword})
	assert.Equal(t, http.StatusUnauthorized, status, "and a signed-in session")
}
