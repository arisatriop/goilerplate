package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TLS modes for email.smtp.tls.
const (
	// SMTPTLSStartTLS connects in plain text and upgrades before authenticating, usually on
	// port 587. The upgrade is required: a server that does not offer it is refused rather
	// than sent credentials in the clear.
	SMTPTLSStartTLS = "starttls"
	// SMTPTLSImplicit speaks TLS from the first byte, usually on port 465.
	SMTPTLSImplicit = "implicit"
	// SMTPTLSNone never encrypts. Only for a relay on the same host or network, or a local
	// catcher such as Mailpit; config validation refuses it in production.
	SMTPTLSNone = "none"
)

// SMTPConfig configures the smtp driver.
type SMTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	// Username and Password authenticate with PLAIN. Leave both empty for a relay that does
	// not authenticate.
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	// TLS is starttls (default), implicit or none.
	TLS string `mapstructure:"tls"`
	// Timeout bounds one delivery, connection included. Default 30s.
	Timeout time.Duration `mapstructure:"timeout"`
}

// TLSOrDefault returns the configured TLS mode in lower case, or SMTPTLSStartTLS when unset.
func (c SMTPConfig) TLSOrDefault() string {
	mode := strings.ToLower(strings.TrimSpace(c.TLS))
	if mode == "" {
		return SMTPTLSStartTLS
	}
	return mode
}

// SMTPSender delivers through an SMTP server, opening one connection per message. Transactional
// volume does not justify a pool, and a pooled connection is one more thing to go stale.
type SMTPSender struct {
	from *mail.Address
	cfg  SMTPConfig
}

// NewSMTPSender returns a sender for cfg. It does not connect until the first Send.
func NewSMTPSender(from *mail.Address, cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{from: from, cfg: cfg}
}

// Send delivers msg.
func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	to, err := msg.validate()
	if err != nil {
		return err
	}

	body, err := buildMIME(s.from, to, msg, time.Now().UTC(),
		"<"+uuid.NewString()+"@"+messageIDDomain(s.from)+">")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, timeoutOrDefault(s.cfg.Timeout))
	defer cancel()

	client, err := s.dial(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }() // after Quit this only releases the socket

	if err := s.deliver(client, to, body); err != nil {
		// A cancelled context closes the connection underneath the client, which surfaces as
		// a network error that says nothing about why.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("smtp: %w", ctxErr)
		}
		return err
	}
	return nil
}

// dial connects and, in starttls mode, upgrades the connection. net/smtp takes no context, so
// the deadline is put on the socket and a cancellation closes it.
func (s *SMTPSender) dial(ctx context.Context) (*smtp.Client, error) {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	tlsConfig := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
	mode := s.cfg.TLSOrDefault()

	var conn net.Conn
	var err error
	if mode == SMTPTLSImplicit {
		conn, err = (&tls.Dialer{Config: tlsConfig}).DialContext(ctx, "tcp", addr)
	} else {
		conn, err = (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return nil, fmt.Errorf("smtp: connecting to %s: %w", addr, err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline) // only fails on a closed connection, which the next read reports
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		stop()
		_ = conn.Close()
		return nil, fmt.Errorf("smtp: greeting: %w", err)
	}

	if mode == SMTPTLSStartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			_ = client.Close()
			return nil, errors.New("smtp: server does not offer STARTTLS; set email.smtp.tls to implicit or none")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("smtp: starttls: %w", err)
		}
	}

	return client, nil
}

func (s *SMTPSender) deliver(client *smtp.Client, to *mail.Address, body []byte) error {
	if s.cfg.Username != "" {
		// PlainAuth refuses to send the password over an unencrypted connection to anything
		// but localhost, which is the protection tls=none would otherwise lack.
		if err := client.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return fmt.Errorf("smtp: auth: %w", err)
		}
	}
	if err := client.Mail(s.from.Address); err != nil {
		return fmt.Errorf("smtp: MAIL FROM: %w", err)
	}
	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("smtp: RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: DATA: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("smtp: writing message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp: message rejected: %w", err)
	}

	// The message is accepted once DATA closes; a failed QUIT does not unsend it.
	_ = client.Quit()
	return nil
}
