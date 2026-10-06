package integration

import (
	"net/http"
	"sync"
	"testing"

	"goilerplate/internal/domain/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireVerification is the stack with auth.require_email_verification on.
func requireVerification(flows *auth.EmailFlows) { flows.RequireVerifiedEmail = true }

// sendCode asks for a verification code for email and returns it from the inbox.
func (s *stack) sendCode(t *testing.T, email string) string {
	t.Helper()

	status, body := s.do(t, http.MethodPost, "/send-verification-email", "", map[string]string{"email": email})
	require.Equal(t, http.StatusOK, status, "send-verification-email: %s", body)

	code := s.inbox.latestCode(email)
	require.NotEmpty(t, code, "no verification code reached %s", email)
	return code
}

func (s *stack) verify(t *testing.T, email, code string) (int, []byte) {
	t.Helper()
	return s.do(t, http.MethodPost, "/verify-email", "", map[string]string{"email": email, "code": code})
}

func otherCode(code string) string {
	if code == "000000" {
		return "000001"
	}
	return "000000"
}

// With verification required, the password alone is not enough to sign in, and the code from the
// inbox is what unlocks it — persisted, so it holds on the next login too.
func TestEmailVerification_RequiredBeforeLogin(t *testing.T) {
	s := newStackWith(t, cacheModes()[0], requireVerification)
	email := s.createUser(t, testPassword)

	status, body := s.do(t, http.MethodPost, "/login", "", loginRequest{Email: email, Password: testPassword})
	assert.Equal(t, http.StatusForbidden, status)
	assertCode(t, body, "email_not_verified")

	status, body = s.do(t, http.MethodPost, "/login", "", loginRequest{Email: email, Password: "not-the-password"})
	assert.Equal(t, http.StatusUnauthorized, status, "without the password, nothing is said about verification")
	assertCode(t, body, "invalid_credentials")

	code := s.sendCode(t, email)
	status, body = s.verify(t, email, code)
	require.Equal(t, http.StatusOK, status, "verify-email: %s", body)

	s.login(t, email, testPassword)

	status, body = s.verify(t, email, code)
	assert.Equal(t, http.StatusBadRequest, status, "a code works once")
	assertCode(t, body, "invalid_verification_code")
}

// The deferred test from the previous roadmap, end to end against the real SQL: after
// max_attempts wrong guesses the code is dead, even when the right one is tried next.
func TestEmailVerification_CodeDiesAfterMaxAttempts(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	email := s.createUser(t, testPassword)
	code := s.sendCode(t, email)

	for range testOTPAttempts {
		status, body := s.verify(t, email, otherCode(code))
		require.Equal(t, http.StatusBadRequest, status)
		assertCode(t, body, "invalid_verification_code")
	}

	status, _ := s.verify(t, email, code)
	assert.Equal(t, http.StatusBadRequest, status, "the right code no longer works")

	fresh := s.sendCode(t, email)
	status, body := s.verify(t, email, fresh)
	assert.Equal(t, http.StatusOK, status, "a new code starts a new count: %s", body)
}

// A password reset proves the user reads the inbox, so it counts as verification. Without it, a
// user who reset before verifying would be refused at the next login.
func TestEmailVerification_PasswordResetVerifiesTheAddress(t *testing.T) {
	s := newStackWith(t, cacheModes()[0], requireVerification)
	email := s.createUser(t, testPassword)

	token := s.forgot(t, email)
	status, body := s.do(t, http.MethodPost, "/reset-password", "", resetRequest{token, newTestPassword})
	require.Equal(t, http.StatusOK, status, "reset-password: %s", body)

	s.login(t, email, newTestPassword)
}

// Guesses sent all at once must not get more tries than guesses sent one by one. The attempt is
// counted in the database before the code is compared, so however the guesses interleave, the
// count reaches the ceiling and the code dies. Guesses arriving after that find no live code
// and are not counted at all.
func TestEmailVerification_ConcurrentGuessesShareOneCeiling(t *testing.T) {
	s := newStack(t, cacheModes()[0])
	email := s.createUser(t, testPassword)
	code := s.sendCode(t, email)

	const guesses = 20
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range guesses {
		wg.Go(func() {
			<-start
			status, _ := s.verify(t, email, otherCode(code))
			assert.Equal(t, http.StatusBadRequest, status)
		})
	}
	close(start)
	wg.Wait()

	var attempts int
	require.NoError(t, s.db.Raw(`SELECT attempts FROM one_time_tokens t JOIN users u ON u.id = t.user_id
		WHERE u.email = ? AND t.token_type = ?`, email, auth.OneTimeTokenEmailVerification).Scan(&attempts).Error)
	assert.GreaterOrEqual(t, attempts, testOTPAttempts, "the burst reached the ceiling")

	status, _ := s.verify(t, email, code)
	assert.Equal(t, http.StatusBadRequest, status, "the ceiling held under concurrency")
}
