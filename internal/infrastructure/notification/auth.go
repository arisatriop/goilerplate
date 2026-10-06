// Package notification turns the domain's notifications into messages a user receives. The
// domain says what happened and hands over what the message must carry; this package decides
// the wording and the links, and pkg/email delivers them.
package notification

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"net/url"
	"strings"
	texttemplate "text/template"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/email"
	"goilerplate/pkg/utils"
)

// AuthMailer is the auth domain's Notifier, delivering by email.
type AuthMailer struct {
	sender   email.Sender
	appName  string
	baseURL  *url.URL
	resetURL *url.URL
}

var _ auth.Notifier = (*AuthMailer)(nil)

// AuthMailerOptions configures an AuthMailer.
type AuthMailerOptions struct {
	// AppName signs the messages and prefixes their subjects.
	AppName string
	// FrontendBaseURL and ResetPasswordPath locate the page a reset link opens.
	FrontendBaseURL   string
	ResetPasswordPath string
}

// NewAuthMailer returns an AuthMailer sending through sender. It fails when the frontend URL
// does not parse, which config validation normally catches first.
func NewAuthMailer(sender email.Sender, opts AuthMailerOptions) (*AuthMailer, error) {
	base, err := url.Parse(strings.TrimSpace(opts.FrontendBaseURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("frontend.base_url %q is not an absolute URL", opts.FrontendBaseURL)
	}

	appName := strings.TrimSpace(opts.AppName)
	if appName == "" {
		appName = "Goilerplate"
	}

	return &AuthMailer{
		sender:   sender,
		appName:  appName,
		baseURL:  base,
		resetURL: base.JoinPath(opts.ResetPasswordPath),
	}, nil
}

// SendPasswordReset emails the reset link.
func (m *AuthMailer) SendPasswordReset(ctx context.Context, notice auth.PasswordResetNotice) error {
	link := *m.resetURL
	link.RawQuery = url.Values{"token": {notice.Token}}.Encode()

	data := passwordResetData{
		AppName:   m.appName,
		Name:      notice.Name,
		Link:      link.String(),
		ExpiresIn: humanDuration(notice.ExpiresAt.Sub(utils.Now())),
	}

	var text, html bytes.Buffer
	if err := passwordResetText.Execute(&text, data); err != nil {
		return fmt.Errorf("rendering password reset text: %w", err)
	}
	if err := passwordResetHTML.Execute(&html, data); err != nil {
		return fmt.Errorf("rendering password reset html: %w", err)
	}

	return m.sender.Send(ctx, email.Message{
		To:      notice.Email,
		Subject: m.appName + ": reset your password",
		Text:    text.String(),
		HTML:    html.String(),
	})
}

// SendEmailVerification emails the verification code.
func (m *AuthMailer) SendEmailVerification(ctx context.Context, notice auth.EmailVerificationNotice) error {
	data := emailVerificationData{
		AppName:   m.appName,
		Name:      notice.Name,
		Code:      notice.Code,
		ExpiresIn: humanDuration(notice.ExpiresAt.Sub(utils.Now())),
	}

	var text, html bytes.Buffer
	if err := emailVerificationText.Execute(&text, data); err != nil {
		return fmt.Errorf("rendering verification text: %w", err)
	}
	if err := emailVerificationHTML.Execute(&html, data); err != nil {
		return fmt.Errorf("rendering verification html: %w", err)
	}

	// The code stays out of the subject: it would show in notification previews and on a
	// locked screen, and the subject is the one part of the message the queue logs.
	return m.sender.Send(ctx, email.Message{
		To:      notice.Email,
		Subject: m.appName + ": verify your email address",
		Text:    text.String(),
		HTML:    html.String(),
	})
}

// SendAccountExists tells the owner of an address that it was registered again.
func (m *AuthMailer) SendAccountExists(ctx context.Context, notice auth.AccountExistsNotice) error {
	return m.render(ctx, notice.Email, "you already have an account", accountExistsText, accountExistsHTML, messageData{
		AppName: m.appName, Name: notice.Name, Link: m.baseURL.String(),
	})
}

// SendEmailChangeCode emails the code confirming a change of address, to the new address.
func (m *AuthMailer) SendEmailChangeCode(ctx context.Context, notice auth.EmailChangeNotice) error {
	return m.render(ctx, notice.NewEmail, "confirm your new email address", emailChangeText, emailChangeHTML, messageData{
		AppName: m.appName, Name: notice.Name, Code: notice.Code,
		ExpiresIn: humanDuration(notice.ExpiresAt.Sub(utils.Now())),
	})
}

// SendEmailChanged tells the old address that the account has moved.
func (m *AuthMailer) SendEmailChanged(ctx context.Context, notice auth.EmailChangedNotice) error {
	return m.render(ctx, notice.OldEmail, "your email address was changed", emailChangedText, emailChangedHTML, messageData{
		AppName: m.appName, Name: notice.Name, NewEmail: notice.NewEmail,
		Link: m.resetURL.String(),
	})
}

// messageData is what the account templates below may use.
type messageData struct {
	AppName   string
	Name      string
	Link      string
	Code      string
	ExpiresIn string
	NewEmail  string
}

// render executes both templates and queues the message.
func (m *AuthMailer) render(ctx context.Context, to, subject string, text *texttemplate.Template, html *htmltemplate.Template, data messageData) error {
	var textBody, htmlBody bytes.Buffer
	if err := text.Execute(&textBody, data); err != nil {
		return fmt.Errorf("rendering %s: %w", text.Name(), err)
	}
	if err := html.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("rendering %s: %w", html.Name(), err)
	}

	return m.sender.Send(ctx, email.Message{
		To:      to,
		Subject: m.appName + ": " + subject,
		Text:    textBody.String(),
		HTML:    htmlBody.String(),
	})
}

type emailVerificationData struct {
	AppName   string
	Name      string
	Code      string
	ExpiresIn string
}

type passwordResetData struct {
	AppName   string
	Name      string
	Link      string
	ExpiresIn string
}

// The HTML version goes through html/template, so a user's display name cannot inject markup
// into an email that carries a credential.
var (
	passwordResetText = texttemplate.Must(texttemplate.New("reset.txt").Parse(`Hi{{if .Name}} {{.Name}}{{end}},

Someone asked to reset the password for your {{.AppName}} account. To choose a new one, open this link:

{{.Link}}

The link works once and expires in {{.ExpiresIn}}. Resetting your password signs you out on every device.

If you did not ask for this, you can ignore this email: your password stays as it is.

— {{.AppName}}
`))

	passwordResetHTML = htmltemplate.Must(htmltemplate.New("reset.html").Parse(`<!doctype html>
<html>
<body style="font-family: sans-serif; line-height: 1.5; color: #222;">
<p>Hi{{if .Name}} {{.Name}}{{end}},</p>
<p>Someone asked to reset the password for your {{.AppName}} account. To choose a new one, use the button below.</p>
<p><a href="{{.Link}}" style="display: inline-block; padding: 10px 18px; background: #222; color: #fff; text-decoration: none; border-radius: 4px;">Reset password</a></p>
<p>Or open this link: <br><a href="{{.Link}}">{{.Link}}</a></p>
<p>The link works once and expires in {{.ExpiresIn}}. Resetting your password signs you out on every device.</p>
<p>If you did not ask for this, you can ignore this email: your password stays as it is.</p>
<p>— {{.AppName}}</p>
</body>
</html>
`))
)

var (
	emailVerificationText = texttemplate.Must(texttemplate.New("verify.txt").Parse(`Hi{{if .Name}} {{.Name}}{{end}},

Your {{.AppName}} verification code is:

{{.Code}}

It expires in {{.ExpiresIn}}. Nobody from {{.AppName}} will ask you for this code.

If you did not create an account, you can ignore this email.

— {{.AppName}}
`))

	emailVerificationHTML = htmltemplate.Must(htmltemplate.New("verify.html").Parse(`<!doctype html>
<html>
<body style="font-family: sans-serif; line-height: 1.5; color: #222;">
<p>Hi{{if .Name}} {{.Name}}{{end}},</p>
<p>Your {{.AppName}} verification code is:</p>
<p style="font-size: 28px; font-weight: bold; letter-spacing: 6px; font-family: monospace;">{{.Code}}</p>
<p>It expires in {{.ExpiresIn}}. Nobody from {{.AppName}} will ask you for this code.</p>
<p>If you did not create an account, you can ignore this email.</p>
<p>— {{.AppName}}</p>
</body>
</html>
`))
)

var (
	accountExistsText = texttemplate.Must(texttemplate.New("exists.txt").Parse(`Hi{{if .Name}} {{.Name}}{{end}},

Someone tried to create a {{.AppName}} account with this email address. You already have one, so nothing was created.

Sign in at {{.Link}}. If you forgot your password, use "Forgot password" there.

If this was not you, you can ignore this email.

— {{.AppName}}
`))

	accountExistsHTML = htmltemplate.Must(htmltemplate.New("exists.html").Parse(`<!doctype html>
<html>
<body style="font-family: sans-serif; line-height: 1.5; color: #222;">
<p>Hi{{if .Name}} {{.Name}}{{end}},</p>
<p>Someone tried to create a {{.AppName}} account with this email address. You already have one, so nothing was created.</p>
<p><a href="{{.Link}}">Sign in</a>. If you forgot your password, use "Forgot password" there.</p>
<p>If this was not you, you can ignore this email.</p>
<p>— {{.AppName}}</p>
</body>
</html>
`))

	emailChangeText = texttemplate.Must(texttemplate.New("change.txt").Parse(`Hi{{if .Name}} {{.Name}}{{end}},

To move your {{.AppName}} account to this email address, enter this code:

{{.Code}}

It expires in {{.ExpiresIn}}. Nobody from {{.AppName}} will ask you for this code.

If you did not ask for this, ignore this email: nothing changes without the code.

— {{.AppName}}
`))

	emailChangeHTML = htmltemplate.Must(htmltemplate.New("change.html").Parse(`<!doctype html>
<html>
<body style="font-family: sans-serif; line-height: 1.5; color: #222;">
<p>Hi{{if .Name}} {{.Name}}{{end}},</p>
<p>To move your {{.AppName}} account to this email address, enter this code:</p>
<p style="font-size: 28px; font-weight: bold; letter-spacing: 6px; font-family: monospace;">{{.Code}}</p>
<p>It expires in {{.ExpiresIn}}. Nobody from {{.AppName}} will ask you for this code.</p>
<p>If you did not ask for this, ignore this email: nothing changes without the code.</p>
<p>— {{.AppName}}</p>
</body>
</html>
`))

	emailChangedText = texttemplate.Must(texttemplate.New("changed.txt").Parse(`Hi{{if .Name}} {{.Name}}{{end}},

The email address of your {{.AppName}} account was changed to {{.NewEmail}}. Emails about your account now go there.

If you did not do this, someone else may be signed in to your account. Reset your password at {{.Link}} — that signs out every device — and contact support.

— {{.AppName}}
`))

	emailChangedHTML = htmltemplate.Must(htmltemplate.New("changed.html").Parse(`<!doctype html>
<html>
<body style="font-family: sans-serif; line-height: 1.5; color: #222;">
<p>Hi{{if .Name}} {{.Name}}{{end}},</p>
<p>The email address of your {{.AppName}} account was changed to <strong>{{.NewEmail}}</strong>. Emails about your account now go there.</p>
<p>If you did not do this, someone else may be signed in to your account. <a href="{{.Link}}">Reset your password</a> — that signs out every device — and contact support.</p>
<p>— {{.AppName}}</p>
</body>
</html>
`))
)

// humanDuration renders a link lifetime for a person: "30 minutes", "1 hour", "2 hours".
// It rounds to the nearest minute, so a token issued a moment ago still reads as its TTL.
func humanDuration(d time.Duration) string {
	minutes := int(d.Round(time.Minute) / time.Minute)
	switch {
	case minutes <= 1:
		return "1 minute"
	case minutes < 60:
		return fmt.Sprintf("%d minutes", minutes)
	case minutes%60 != 0:
		return fmt.Sprintf("%d minutes", minutes)
	case minutes == 60:
		return "1 hour"
	default:
		return fmt.Sprintf("%d hours", minutes/60)
	}
}
