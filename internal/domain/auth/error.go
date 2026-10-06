package auth

import (
	"errors"

	"goilerplate/pkg/apperr"
	"goilerplate/pkg/constants"
)

// ErrNotFound is returned by the repository when an update or delete matches no record.
var ErrNotFound = errors.New("auth: record not found")

// MsgSessionNotFound is the message of ErrSessionNotFound.
const MsgSessionNotFound = "Session not found"

// Client errors of the auth domain. Each carries a stable code; the delivery layer decides the
// status. The kinds are chosen so that nothing distinguishes cases an attacker must not be able
// to tell apart: an unknown email and a wrong password are both ErrInvalidCredentials.
var (
	// ErrUnauthorized: the credential or session is not (or no longer) acceptable. One answer
	// for every reason, so the response never says which check failed.
	ErrUnauthorized = apperr.New(apperr.Unauthenticated, "unauthorized", constants.MsgUnauthorized)
	// ErrInvalidCredentials: login or re-authentication with a wrong email or password.
	ErrInvalidCredentials = apperr.New(apperr.Unauthenticated, "invalid_credentials", constants.MsgInvalidCredential)
	// ErrAccountLocked: too many failed logins. Returned whether or not the password was right.
	ErrAccountLocked = apperr.New(apperr.Unauthenticated, "account_locked", constants.MsgAccountLocked)
	// ErrAccountDisabled: the account exists and the credentials were right, but it is inactive.
	ErrAccountDisabled = apperr.New(apperr.Forbidden, "account_disabled", constants.MsgAccountDisabled)
	// ErrEmailAlreadyRegistered: registration with an address that has an account.
	ErrEmailAlreadyRegistered = apperr.New(apperr.Conflict, "email_already_registered", "Email is already registered")
	// ErrSessionNotFound: the session does not exist, is not the caller's, or is already revoked
	// — deliberately indistinguishable, so another user's session IDs cannot be confirmed.
	ErrSessionNotFound = apperr.New(apperr.NotFound, "session_not_found", MsgSessionNotFound)
	// ErrInvalidResetToken: the reset link is unknown, used, superseded or expired — one answer
	// for all of them, so a guessed token learns nothing about which tokens exist.
	ErrInvalidResetToken = apperr.New(apperr.Invalid, "invalid_reset_token", "Reset link is invalid or has expired")
	// ErrInvalidVerificationCode: wrong, expired, superseded or exhausted code, or no code for
	// that address — one answer, so guessing teaches nothing but "no".
	ErrInvalidVerificationCode = apperr.New(apperr.Invalid, "invalid_verification_code", "Verification code is invalid or has expired")
	// ErrEmailNotVerified: correct credentials, but auth.require_email_verification is on and
	// the address is not verified yet. Like ErrAccountDisabled, it is only ever said to someone
	// who just proved they know the password.
	ErrEmailNotVerified = apperr.New(apperr.Forbidden, "email_not_verified", "Verify your email address before signing in")
)
