package email

import (
	"context"
	"log/slog"
	"net/mail"
)

// LogSender writes each message to the log instead of delivering it, so every email flow can
// be exercised locally with no mail server.
//
// The body is logged in full, because reading the link out of it is the point. That makes it a
// credential leak anywhere logs are shipped or shared: a reset link in a log is an account
// takeover for anyone who can read the log. Config validation refuses this driver in production.
type LogSender struct {
	from   *mail.Address
	logger *slog.Logger
}

// NewLogSender returns a LogSender writing to logger, or to slog.Default when logger is nil.
func NewLogSender(from *mail.Address, logger *slog.Logger) *LogSender {
	return &LogSender{from: from, logger: logger}
}

// Send logs msg.
func (s *LogSender) Send(ctx context.Context, msg Message) error {
	to, err := msg.validate()
	if err != nil {
		return err
	}

	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.InfoContext(ctx, "email not delivered: log driver",
		"from", formatAddress(s.from),
		"to", formatAddress(to),
		"subject", msg.Subject,
		"text", msg.Text,
	)
	return nil
}
