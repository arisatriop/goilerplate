// Package email sends transactional email through a driver chosen in configuration:
//
//	log     writes the message to the log instead of sending it (development only)
//	smtp    any SMTP server or relay: Mailpit locally, or a provider's SMTP endpoint
//	ses     Amazon SES through its API, so credentials can come from an IAM role
//	resend  Resend through its HTTP API
//
// Callers depend on Sender and never on a driver, so changing provider is a config change.
// Queue wraps any Sender to deliver in the background.
package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"
)

// Driver names for email.driver.
const (
	DriverLog    = "log"
	DriverSMTP   = "smtp"
	DriverSES    = "ses"
	DriverResend = "resend"
)

// DefaultSendTimeout bounds one delivery attempt when a driver's timeout is unset.
const DefaultSendTimeout = 30 * time.Second

// Config selects and configures the driver. Only the block matching Driver is read.
type Config struct {
	// Driver is log (default), smtp, ses or resend.
	Driver string `mapstructure:"driver"`
	// From is the sender, as an address or "Name <address>". Providers only accept a sender
	// on a domain you have verified with them.
	From   string       `mapstructure:"from"`
	SMTP   SMTPConfig   `mapstructure:"smtp"`
	SES    SESConfig    `mapstructure:"ses"`
	Resend ResendConfig `mapstructure:"resend"`
}

// DriverOrDefault returns the configured driver in lower case, or DriverLog when unset.
func (c Config) DriverOrDefault() string {
	driver := strings.ToLower(strings.TrimSpace(c.Driver))
	if driver == "" {
		return DriverLog
	}
	return driver
}

// Message is one email to one recipient.
type Message struct {
	To      string
	Subject string
	// Text is the plain-text body. It is always sent: some clients show nothing else, and a
	// message with no text part scores worse with spam filters.
	Text string
	// HTML is an optional alternative to Text.
	HTML string
}

// Sender delivers a message. Implementations must honour ctx cancellation.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// ErrInvalidMessage is returned for a message no driver could send. It is a programming error
// on the caller's side, never a provider failure, so retrying it is pointless.
var ErrInvalidMessage = errors.New("email: invalid message")

// New builds the Sender for cfg.Driver. ctx is used only while building clients.
func New(ctx context.Context, cfg Config, logger *slog.Logger) (Sender, error) {
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("parsing email.from: %w", err)
	}

	switch cfg.DriverOrDefault() {
	case DriverLog:
		return NewLogSender(from, logger), nil
	case DriverSMTP:
		return NewSMTPSender(from, cfg.SMTP), nil
	case DriverSES:
		return NewSESSender(ctx, from, cfg.SES)
	case DriverResend:
		return NewResendSender(from, cfg.Resend), nil
	default:
		return nil, fmt.Errorf("unknown email driver %q", cfg.Driver)
	}
}

// validate checks what every driver relies on. A CR or LF in the subject is refused rather
// than encoded: it is either a bug or an attempt to inject a header, and neither should reach
// a mail server.
func (m Message) validate() (*mail.Address, error) {
	to, err := mail.ParseAddress(m.To)
	if err != nil {
		return nil, fmt.Errorf("%w: recipient: %w", ErrInvalidMessage, err)
	}
	if strings.ContainsAny(m.Subject, "\r\n") {
		return nil, fmt.Errorf("%w: subject contains a line break", ErrInvalidMessage)
	}
	if strings.TrimSpace(m.Subject) == "" {
		return nil, fmt.Errorf("%w: subject is empty", ErrInvalidMessage)
	}
	if strings.TrimSpace(m.Text) == "" {
		return nil, fmt.Errorf("%w: text body is empty", ErrInvalidMessage)
	}
	return to, nil
}

func timeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return DefaultSendTimeout
}

// formatAddress renders an address the way a provider expects it: bare when there is no
// display name, rather than mail.Address.String's "<address>".
func formatAddress(address *mail.Address) string {
	if address.Name == "" {
		return address.Address
	}
	return address.String()
}
