package notification

import (
	"context"
	"net/url"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/email"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturingSender struct{ sent []email.Message }

func (s *capturingSender) Send(_ context.Context, msg email.Message) error {
	s.sent = append(s.sent, msg)
	return nil
}

func TestAuthMailer_PasswordResetLinksToTheFrontend(t *testing.T) {
	sender := &capturingSender{}
	mailer, err := NewAuthMailer(sender, AuthMailerOptions{
		AppName:           "Acme",
		FrontendBaseURL:   "https://app.example.com/portal/",
		ResetPasswordPath: "/reset-password",
	})
	require.NoError(t, err)

	// The token is base64url with padding, so it carries a "=" that must be escaped in a query.
	const token = "abc-_DEF123="
	err = mailer.SendPasswordReset(t.Context(), auth.PasswordResetNotice{
		Email:     "user@example.org",
		Name:      "Ana",
		Token:     token,
		ExpiresAt: utils.Now().Add(30 * time.Minute),
	})
	require.NoError(t, err)

	require.Len(t, sender.sent, 1)
	msg := sender.sent[0]
	assert.Equal(t, "user@example.org", msg.To)
	assert.Equal(t, "Acme: reset your password", msg.Subject)

	const wantLink = "https://app.example.com/portal/reset-password?token=abc-_DEF123%3D"
	assert.Contains(t, msg.Text, wantLink)
	assert.Contains(t, msg.Text, "expires in 30 minutes")
	assert.Contains(t, msg.HTML, `href="`+wantLink+`"`)

	parsed, err := url.Parse(wantLink)
	require.NoError(t, err)
	assert.Equal(t, token, parsed.Query().Get("token"), "the frontend reads back exactly the token issued")
}

func TestAuthMailer_EscapesTheNameInHTML(t *testing.T) {
	sender := &capturingSender{}
	mailer, err := NewAuthMailer(sender, AuthMailerOptions{FrontendBaseURL: "https://app.example.com", ResetPasswordPath: "/reset-password"})
	require.NoError(t, err)

	err = mailer.SendPasswordReset(t.Context(), auth.PasswordResetNotice{
		Email:     "user@example.org",
		Name:      `<a href="https://evil.example">click</a>`,
		Token:     "t",
		ExpiresAt: utils.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	html := sender.sent[0].HTML
	assert.NotContains(t, html, `<a href="https://evil.example">`)
	assert.Contains(t, html, "&lt;a href=")
	assert.Contains(t, html, "https://app.example.com/reset-password?token=t")
}

func TestNewAuthMailer_RefusesARelativeBaseURL(t *testing.T) {
	_, err := NewAuthMailer(&capturingSender{}, AuthMailerOptions{FrontendBaseURL: "app.example.com"})

	assert.Error(t, err)
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{30*time.Minute - 2*time.Second, "30 minutes"},
		{time.Minute, "1 minute"},
		{0, "1 minute"},
		{time.Hour, "1 hour"},
		{90 * time.Minute, "90 minutes"},
		{24 * time.Hour, "24 hours"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, humanDuration(tt.in))
		})
	}
}

func TestAuthMailer_EmailVerificationCarriesTheCodeInTheBodyOnly(t *testing.T) {
	sender := &capturingSender{}
	mailer, err := NewAuthMailer(sender, AuthMailerOptions{AppName: "Acme", FrontendBaseURL: "https://app.example.com"})
	require.NoError(t, err)

	err = mailer.SendEmailVerification(t.Context(), auth.EmailVerificationNotice{
		Email:     "user@example.org",
		Name:      "Ana",
		Code:      "042917",
		ExpiresAt: utils.Now().Add(15 * time.Minute),
	})
	require.NoError(t, err)

	require.Len(t, sender.sent, 1)
	msg := sender.sent[0]
	assert.Equal(t, "user@example.org", msg.To)
	assert.Equal(t, "Acme: verify your email address", msg.Subject)
	assert.NotContains(t, msg.Subject, "042917", "a code in the subject shows on a locked screen")
	assert.Contains(t, msg.Text, "042917", "leading zeros survive")
	assert.Contains(t, msg.Text, "expires in 15 minutes")
	assert.Contains(t, msg.HTML, "042917")
}

func TestAuthMailer_AccountMessagesGoToTheRightInbox(t *testing.T) {
	sender := &capturingSender{}
	mailer, err := NewAuthMailer(sender, AuthMailerOptions{
		AppName: "Acme", FrontendBaseURL: "https://app.example.com", ResetPasswordPath: "/reset-password",
	})
	require.NoError(t, err)
	ctx := t.Context()

	require.NoError(t, mailer.SendAccountExists(ctx, auth.AccountExistsNotice{Email: "ana@example.org", Name: "Ana"}))
	require.NoError(t, mailer.SendEmailChangeCode(ctx, auth.EmailChangeNotice{
		NewEmail: "ana.new@example.org", Name: "Ana", Code: "042917", ExpiresAt: utils.Now().Add(15 * time.Minute),
	}))
	require.NoError(t, mailer.SendEmailChanged(ctx, auth.EmailChangedNotice{
		OldEmail: "ana@example.org", NewEmail: "ana.new@example.org", Name: "Ana",
	}))

	require.Len(t, sender.sent, 3)

	exists := sender.sent[0]
	assert.Equal(t, "ana@example.org", exists.To)
	assert.Equal(t, "Acme: you already have an account", exists.Subject)
	assert.Contains(t, exists.Text, "https://app.example.com")

	change := sender.sent[1]
	assert.Equal(t, "ana.new@example.org", change.To, "the code proves the user reads the new inbox")
	assert.Contains(t, change.Text, "042917")
	assert.NotContains(t, change.Subject, "042917")

	changed := sender.sent[2]
	assert.Equal(t, "ana@example.org", changed.To, "the old inbox is warned")
	assert.Contains(t, changed.Text, "ana.new@example.org")
	assert.Contains(t, changed.Text, "https://app.example.com/reset-password")
	assert.Contains(t, changed.HTML, "<strong>ana.new@example.org</strong>")
}
