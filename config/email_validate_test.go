package config

import (
	"testing"
	"time"

	"goilerplate/pkg/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withEmail is a valid configuration with the email flows on, using the given driver.
func withEmail(env string, cfg email.Config) *Config {
	c := validConfig()
	c.App.Env = env
	c.Auth.Email.Enabled = true
	c.Email = cfg
	c.Frontend = Frontend{BaseURL: "https://app.example.com"}
	c.Auth.OTP.Secret = "Qm8vT2xZr5WcN1kJ7pHd3sFg6aL9eB4u"
	return c
}

func smtpConfig() email.Config {
	return email.Config{
		Driver: email.DriverSMTP,
		From:   "Goilerplate <no-reply@example.com>",
		SMTP:   email.SMTPConfig{Host: "smtp.example.com", Port: 587, Username: "mailer", Password: "a-real-smtp-password"},
	}
}

func TestValidate_EmailAcceptsEachDriver(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
	}{
		{"log locally", withEmail("local", email.Config{From: "no-reply@example.com"})},
		{"smtp with starttls", withEmail(envProduction, smtpConfig())},
		{"smtp relay without auth", withEmail("dev", email.Config{
			Driver: "smtp", From: "no-reply@example.com",
			SMTP: email.SMTPConfig{Host: "mailpit", Port: 1025, TLS: "none"},
		})},
		{"ses on the default credential chain", withEmail(envProduction, email.Config{
			Driver: "ses", From: "no-reply@example.com", SES: email.SESConfig{Region: "ap-southeast-1"},
		})},
		{"resend", withEmail(envProduction, email.Config{
			Driver: "resend", From: "no-reply@example.com", Resend: email.ResendConfig{APIKey: "re_9f1c2e7a4b3d4a8e9c612d5f"},
		})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.IsProduction() {
				tt.cfg.InternalAuth = InternalAuth{Mode: InternalAuthSharedSecret, Secret: validAccessSecret}
			}
			assert.NoError(t, tt.cfg.Validate())
		})
	}
}

func TestValidate_EmailRules(t *testing.T) {
	tests := []struct {
		name    string
		cfg     func() *Config
		wantErr string
	}{
		{"log driver in production", func() *Config {
			return withEmail(envProduction, email.Config{From: "no-reply@example.com"})
		}, "email.driver=log writes reset links into the logs"},
		{"unknown driver", func() *Config {
			return withEmail("dev", email.Config{Driver: "pigeon", From: "no-reply@example.com"})
		}, `email.driver must be log, smtp, ses, or resend, got "pigeon"`},
		{"missing from", func() *Config { return withEmail("dev", email.Config{}) }, "email.from is required"},
		{"malformed from", func() *Config {
			return withEmail("dev", email.Config{From: "no-reply"})
		}, "email.from must be an address"},
		{"smtp without host", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Email.SMTP.Host = ""
			return c
		}, "email.smtp.host is required"},
		{"smtp without port", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Email.SMTP.Port = 0
			return c
		}, "email.smtp.port must be between 1 and 65535"},
		{"smtp username without password", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Email.SMTP.Password = ""
			return c
		}, "email.smtp.password is required"},
		{"smtp unknown tls mode", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Email.SMTP.TLS = "ssl"
			return c
		}, `email.smtp.tls must be starttls, implicit, or none, got "ssl"`},
		{"smtp without tls in production", func() *Config {
			c := withEmail(envProduction, smtpConfig())
			c.Email.SMTP.TLS = "none"
			return c
		}, "email.smtp.tls=none sends mail"},
		{"ses with half a key pair", func() *Config {
			return withEmail("dev", email.Config{Driver: "ses", From: "no-reply@example.com", SES: email.SESConfig{AccessKeyID: "AKID"}})
		}, "must be set together"},
		{"resend without key", func() *Config {
			return withEmail("dev", email.Config{Driver: "resend", From: "no-reply@example.com"})
		}, "email.resend.api_key is required"},
		{"resend with sample key in production", func() *Config {
			return withEmail(envProduction, email.Config{Driver: "resend", From: "no-reply@example.com", Resend: email.ResendConfig{APIKey: "your_resend_key"}})
		}, "email.resend.api_key"},
		{"missing frontend url", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Frontend.BaseURL = ""
			return c
		}, "frontend.base_url is required"},
		{"relative frontend url", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Frontend.BaseURL = "app.example.com"
			return c
		}, "frontend.base_url must be an absolute http(s) URL"},
		{"plain http frontend in production", func() *Config {
			c := withEmail(envProduction, smtpConfig())
			c.Frontend.BaseURL = "http://app.example.com"
			return c
		}, "frontend.base_url must use https in production"},
		{"reset path without leading slash", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Frontend.ResetPasswordPath = "reset"
			return c
		}, "frontend.reset_password_path must start with /"},
		{"negative reset ttl", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Auth.PasswordReset.TTL = -time.Minute
			return c
		}, "auth.password_reset.ttl must not be negative"},
		{"missing otp secret", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Auth.OTP.Secret = ""
			return c
		}, "auth.otp.secret is required"},
		{"short otp secret", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Auth.OTP.Secret = "too-short"
			return c
		}, "auth.otp.secret must be at least 32 bytes"},
		{"sample otp secret in production", func() *Config {
			c := withEmail(envProduction, smtpConfig())
			c.Auth.OTP.Secret = "changeme-changeme-changeme-changeme"
			return c
		}, "auth.otp.secret looks like an example value"},
		{"verification required without email", func() *Config {
			c := validConfig()
			c.Auth.RequireEmailVerification = true
			return c
		}, "auth.require_email_verification requires auth.email.enabled=true"},
		{"negative otp attempts", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Auth.OTP.MaxAttempts = -1
			return c
		}, "auth.otp.max_attempts must not be negative"},
		{"otp cooldown as long as the code", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Auth.OTP = OTP{Secret: c.Auth.OTP.Secret, TTL: time.Minute, ResendCooldown: time.Minute}
			return c
		}, "auth.otp.resend_cooldown must be shorter than auth.otp.ttl"},
		{"cooldown as long as the link", func() *Config {
			c := withEmail("dev", smtpConfig())
			c.Auth.PasswordReset = PasswordReset{TTL: 10 * time.Minute, ResendCooldown: 10 * time.Minute}
			return c
		}, "resend_cooldown must be shorter than auth.password_reset.ttl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg().Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// A deployment that never sends mail must not be asked to configure a provider.
func TestValidate_EmailIsNotCheckedWhenDisabled(t *testing.T) {
	c := validConfig()
	c.Email = email.Config{Driver: "pigeon"}

	assert.NoError(t, c.Validate())
}

func TestPasswordReset_Defaults(t *testing.T) {
	assert.Equal(t, DefaultPasswordResetTTL, PasswordReset{}.TTLOrDefault())
	assert.Equal(t, DefaultPasswordResetResendCooldown, PasswordReset{}.ResendCooldownOrDefault())
	assert.Equal(t, time.Hour, PasswordReset{TTL: time.Hour}.TTLOrDefault())
	assert.Equal(t, "/reset-password", Frontend{}.ResetPasswordPathOrDefault())
	assert.Equal(t, "/auth/reset", Frontend{ResetPasswordPath: "/auth/reset"}.ResetPasswordPathOrDefault())
	assert.Equal(t, DefaultOTPTTL, OTP{}.TTLOrDefault())
	assert.Equal(t, DefaultOTPMaxAttempts, OTP{}.MaxAttemptsOrDefault())
	assert.Equal(t, DefaultOTPResendCooldown, OTP{}.ResendCooldownOrDefault())
	assert.Equal(t, 3, OTP{MaxAttempts: 3}.MaxAttemptsOrDefault())
}
