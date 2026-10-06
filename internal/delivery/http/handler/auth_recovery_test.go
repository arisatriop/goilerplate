package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/password"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryUsecase records what the recovery handlers passed on; any other method panics.
type recoveryUsecase struct {
	auth.Usecase
	err error

	forgotEmail  string
	forgotOrigin auth.RequestOrigin
	resetToken   string
	resetCalled  bool
	sendEmail    string
	verifyCode   string
	verifyCalled bool
}

func (u *recoveryUsecase) SendEmailVerification(_ context.Context, email string, _ auth.RequestOrigin) error {
	u.sendEmail = email
	return u.err
}

func (u *recoveryUsecase) VerifyEmail(_ context.Context, _, code string) error {
	u.verifyCalled, u.verifyCode = true, code
	return u.err
}

func (u *recoveryUsecase) ForgotPassword(_ context.Context, email string, origin auth.RequestOrigin) error {
	u.forgotEmail, u.forgotOrigin = email, origin
	return u.err
}

func (u *recoveryUsecase) ResetPassword(_ context.Context, token, _ string) error {
	u.resetCalled, u.resetToken = true, token
	return u.err
}

func newRecoveryApp(usecase auth.Usecase) *fiber.App {
	h := handler.NewAuth(nil, response.NewValidator(), nil, usecase, nil)
	app := fiber.New()
	app.Post("/auth/forgot-password", h.ForgotPassword)
	app.Post("/auth/reset-password", h.ResetPassword)
	app.Post("/auth/send-verification-email", h.SendVerificationEmail)
	app.Post("/auth/verify-email", h.VerifyEmail)
	return app
}

type envelope struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func postJSON(t *testing.T, app *fiber.App, path, body string, headers map[string]string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(raw)
}

func decodeEnvelope(t *testing.T, body string) envelope {
	t.Helper()
	var env envelope
	require.NoError(t, json.Unmarshal([]byte(body), &env))
	return env
}

func TestAuthForgotPassword_AnswersWithTheGenericMessage(t *testing.T) {
	t.Parallel()
	usecase := &recoveryUsecase{}
	app := newRecoveryApp(usecase)

	status, body := postJSON(t, app, "/auth/forgot-password", `{"email":"ana@example.org"}`,
		map[string]string{fiber.HeaderUserAgent: "Mozilla/5.0"})

	assert.Equal(t, http.StatusOK, status)
	env := decodeEnvelope(t, body)
	assert.True(t, env.Success)
	assert.Equal(t, handler.MsgPasswordResetRequested, env.Message,
		"the message must not say whether an email was sent")
	assert.Equal(t, "ana@example.org", usecase.forgotEmail)
	assert.Equal(t, "Mozilla/5.0", usecase.forgotOrigin.UserAgent)
	assert.NotEmpty(t, usecase.forgotOrigin.IPAddress)
}

func TestAuthForgotPassword_MalformedEmailIs400(t *testing.T) {
	t.Parallel()
	usecase := &recoveryUsecase{}
	app := newRecoveryApp(usecase)

	status, body := postJSON(t, app, "/auth/forgot-password", `{"email":"not-an-email"}`, nil)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "validation_failed", decodeEnvelope(t, body).Code)
	assert.Empty(t, usecase.forgotEmail, "a rejected request must not reach the use case")
}

func TestAuthForgotPassword_UnexpectedFailureIsAGeneric500(t *testing.T) {
	t.Parallel()
	app := newRecoveryApp(&recoveryUsecase{err: errors.New("dial tcp 10.0.0.5:5432: connection refused")})

	status, body := postJSON(t, app, "/auth/forgot-password", `{"email":"ana@example.org"}`, nil)

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NotContains(t, body, "10.0.0.5", "internal error text never reaches the client")
}

func TestAuthResetPassword_Succeeds(t *testing.T) {
	t.Parallel()
	usecase := &recoveryUsecase{}
	app := newRecoveryApp(usecase)

	status, body := postJSON(t, app, "/auth/reset-password",
		`{"token":"abc-_DEF123=","newPassword":"a-new-strong-password"}`, nil)

	assert.Equal(t, http.StatusOK, status)
	assert.True(t, decodeEnvelope(t, body).Success)
	assert.Equal(t, "abc-_DEF123=", usecase.resetToken)
}

func TestAuthResetPassword_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		usecaseErr error
		wantStatus int
		wantCode   string
		reachesUC  bool
	}{
		{"invalid or expired token is 400 with its own code", `{"token":"used","newPassword":"a-new-strong-password"}`,
			auth.ErrInvalidResetToken, http.StatusBadRequest, "invalid_reset_token", true},
		{"password the policy refuses is 400", `{"token":"t","newPassword":"aaaaaaaa"}`,
			password.ErrTooCommon, http.StatusBadRequest, "", true},
		{"missing token is a validation failure", `{"newPassword":"a-new-strong-password"}`,
			nil, http.StatusBadRequest, "validation_failed", false},
		{"oversized token is a validation failure", `{"token":"` + strings.Repeat("x", 257) + `","newPassword":"a-new-strong-password"}`,
			nil, http.StatusBadRequest, "validation_failed", false},
		{"malformed JSON", `{"token":`, nil, http.StatusBadRequest, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			usecase := &recoveryUsecase{err: tt.usecaseErr}
			app := newRecoveryApp(usecase)

			status, body := postJSON(t, app, "/auth/reset-password", tt.body, nil)

			assert.Equal(t, tt.wantStatus, status)
			assert.False(t, decodeEnvelope(t, body).Success)
			if tt.wantCode != "" {
				assert.Equal(t, tt.wantCode, decodeEnvelope(t, body).Code)
			}
			assert.Equal(t, tt.reachesUC, usecase.resetCalled)
		})
	}
}

// Same promise as the change-password route: a rejected password never comes back in the body.
func TestAuthResetPassword_RejectionDoesNotEchoThePassword(t *testing.T) {
	t.Parallel()
	app := newRecoveryApp(&recoveryUsecase{})

	status, body := postJSON(t, app, "/auth/reset-password", `{"token":"t","newPassword":"hunter2"}`, nil)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotContains(t, body, "hunter2")
	assert.Contains(t, body, `"field":"newPassword"`)
}

func TestAuthSendVerificationEmail_AnswersWithTheGenericMessage(t *testing.T) {
	t.Parallel()
	usecase := &recoveryUsecase{}
	app := newRecoveryApp(usecase)

	status, body := postJSON(t, app, "/auth/send-verification-email", `{"email":"ana@example.org"}`, nil)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, handler.MsgVerificationEmailRequested, decodeEnvelope(t, body).Message)
	assert.Equal(t, "ana@example.org", usecase.sendEmail)
}

func TestAuthVerifyEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		usecaseErr error
		wantStatus int
		wantCode   string
		reachesUC  bool
	}{
		{"right code", `{"email":"ana@example.org","code":"042917"}`, nil, http.StatusOK, "", true},
		{"wrong or expired code", `{"email":"ana@example.org","code":"123456"}`,
			auth.ErrInvalidVerificationCode, http.StatusBadRequest, "invalid_verification_code", true},
		{"five digits", `{"email":"ana@example.org","code":"12345"}`, nil, http.StatusBadRequest, "validation_failed", false},
		{"letters", `{"email":"ana@example.org","code":"12345a"}`, nil, http.StatusBadRequest, "validation_failed", false},
		{"missing email", `{"code":"123456"}`, nil, http.StatusBadRequest, "validation_failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			usecase := &recoveryUsecase{err: tt.usecaseErr}
			app := newRecoveryApp(usecase)

			status, body := postJSON(t, app, "/auth/verify-email", tt.body, nil)

			assert.Equal(t, tt.wantStatus, status)
			if tt.wantCode != "" {
				assert.Equal(t, tt.wantCode, decodeEnvelope(t, body).Code)
			}
			assert.Equal(t, tt.reachesUC, usecase.verifyCalled)
			if tt.reachesUC {
				assert.Equal(t, strings.Split(strings.Split(tt.body, `"code":"`)[1], `"`)[0], usecase.verifyCode,
					"leading zeros reach the use case")
			}
		})
	}
}
