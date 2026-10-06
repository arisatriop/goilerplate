package email_test

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"goilerplate/pkg/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSMTP is just enough of an SMTP server to accept one message per connection and record
// the conversation. Behaviour switches cover what the client must react to.
type fakeSMTP struct {
	addr string

	advertiseAuth bool
	hang          bool // accept the connection and never greet

	mu       sync.Mutex
	from     string
	rcpt     string
	data     string
	authLine string
	done     chan struct{}
}

func newFakeSMTP(t *testing.T, configure func(*fakeSMTP)) *fakeSMTP {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &fakeSMTP{addr: listener.Addr().String(), done: make(chan struct{})}
	if configure != nil {
		configure(server)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		server.serve(conn)
	}()

	t.Cleanup(func() {
		_ = listener.Close()
		wg.Wait()
	})
	return server
}

func (s *fakeSMTP) serve(conn net.Conn) {
	defer close(s.done)
	if s.hang {
		_, _ = io.Copy(io.Discard, conn) // returns once the client gives up and closes
		return
	}

	r := bufio.NewReader(conn)
	reply := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }

	reply("220 fake ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		verb := strings.ToUpper(strings.SplitN(line, " ", 2)[0])

		s.mu.Lock()
		switch verb {
		case "EHLO":
			if s.advertiseAuth {
				reply("250-fake")
				reply("250 AUTH PLAIN")
			} else {
				reply("250 fake")
			}
		case "AUTH":
			s.authLine = line
			reply("235 ok")
		case "MAIL":
			s.from = line
			reply("250 ok")
		case "RCPT":
			s.rcpt = line
			reply("250 ok")
		case "DATA":
			reply("354 go ahead")
			var body strings.Builder
			for {
				dataLine, err := r.ReadString('\n')
				if err != nil {
					s.mu.Unlock()
					return
				}
				if dataLine == ".\r\n" {
					break
				}
				body.WriteString(dataLine)
			}
			s.data = body.String()
			reply("250 queued")
		case "QUIT":
			reply("221 bye")
			s.mu.Unlock()
			return
		default:
			reply("502 not implemented")
		}
		s.mu.Unlock()
	}
}

func (s *fakeSMTP) config(t *testing.T) email.SMTPConfig {
	t.Helper()
	host, port, err := net.SplitHostPort(s.addr)
	require.NoError(t, err)
	portNumber, err := strconv.Atoi(port)
	require.NoError(t, err)
	return email.SMTPConfig{Host: host, Port: portNumber, TLS: email.SMTPTLSNone}
}

func mustAddress(t *testing.T, raw string) *mail.Address {
	t.Helper()
	address, err := mail.ParseAddress(raw)
	require.NoError(t, err)
	return address
}

func TestSMTPSender_DeliversAMultipartMessage(t *testing.T) {
	server := newFakeSMTP(t, nil)
	sender := email.NewSMTPSender(mustAddress(t, "Goilerplate <no-reply@example.com>"), server.config(t))

	err := sender.Send(t.Context(), email.Message{
		To:      "user@example.org",
		Subject: "Réinitialiser le mot de passe",
		Text:    "Open https://app.example.com/reset?token=abc",
		HTML:    `<a href="https://app.example.com/reset?token=abc">Reset</a>`,
	})
	require.NoError(t, err)
	<-server.done

	server.mu.Lock()
	defer server.mu.Unlock()
	assert.Equal(t, "MAIL FROM:<no-reply@example.com>", server.from)
	assert.Equal(t, "RCPT TO:<user@example.org>", server.rcpt)

	msg, err := mail.ReadMessage(strings.NewReader(server.data))
	require.NoError(t, err)
	subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	require.NoError(t, err)
	assert.Equal(t, "Réinitialiser le mot de passe", subject, "a non-ASCII subject survives encoding")
	assert.NotEmpty(t, msg.Header.Get("Message-ID"))
	assert.Contains(t, msg.Header.Get("Message-ID"), "@example.com>")

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/alternative", mediaType)

	parts := multipart.NewReader(msg.Body, params["boundary"])
	var types []string
	for {
		part, err := parts.NextPart() // decodes quoted-printable transparently
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		body, err := io.ReadAll(part)
		require.NoError(t, err)
		types = append(types, part.Header.Get("Content-Type"))
		assert.Contains(t, string(body), "https://app.example.com/reset?token=abc",
			"the link must survive quoted-printable intact")
	}
	assert.Equal(t, []string{`text/plain; charset="utf-8"`, `text/html; charset="utf-8"`}, types,
		"plain text first: the last alternative is the preferred one")
}

func TestSMTPSender_AuthenticatesWhenUsernameIsSet(t *testing.T) {
	server := newFakeSMTP(t, func(s *fakeSMTP) { s.advertiseAuth = true })
	cfg := server.config(t)
	cfg.Username, cfg.Password = "mailer", "s3cret"

	err := email.NewSMTPSender(mustAddress(t, "no-reply@example.com"), cfg).
		Send(t.Context(), email.Message{To: "user@example.org", Subject: "Hi", Text: "Hello"})
	require.NoError(t, err)
	<-server.done

	server.mu.Lock()
	defer server.mu.Unlock()
	require.True(t, strings.HasPrefix(server.authLine, "AUTH PLAIN "), server.authLine)
	credentials, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(server.authLine, "AUTH PLAIN "))
	require.NoError(t, err)
	assert.Equal(t, "\x00mailer\x00s3cret", string(credentials))
}

// starttls is the default, and it must be required rather than attempted: falling back to
// plain text would hand the password to anyone on the path who strips the STARTTLS offer.
func TestSMTPSender_StartTLSIsRequiredNotOpportunistic(t *testing.T) {
	server := newFakeSMTP(t, nil) // never offers STARTTLS
	cfg := server.config(t)
	cfg.TLS = ""

	err := email.NewSMTPSender(mustAddress(t, "no-reply@example.com"), cfg).
		Send(t.Context(), email.Message{To: "user@example.org", Subject: "Hi", Text: "Hello"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "STARTTLS")
	<-server.done
	server.mu.Lock()
	defer server.mu.Unlock()
	assert.Empty(t, server.from, "nothing is sent over a connection that could not be upgraded")
}

func TestSMTPSender_GivesUpOnAServerThatNeverAnswers(t *testing.T) {
	server := newFakeSMTP(t, func(s *fakeSMTP) { s.hang = true })
	cfg := server.config(t)
	cfg.Timeout = 100 * time.Millisecond

	start := time.Now()
	err := email.NewSMTPSender(mustAddress(t, "no-reply@example.com"), cfg).
		Send(context.Background(), email.Message{To: "user@example.org", Subject: "Hi", Text: "Hello"})

	require.Error(t, err)
	assert.Less(t, time.Since(start), 5*time.Second, "the timeout bounds the whole delivery")
}

func TestMessage_RejectedBeforeAnythingIsSent(t *testing.T) {
	tests := []struct {
		name string
		msg  email.Message
	}{
		{"header injection through the subject", email.Message{To: "user@example.org", Subject: "Hi\r\nBcc: victim@example.org", Text: "x"}},
		{"malformed recipient", email.Message{To: "not an address", Subject: "Hi", Text: "x"}},
		{"recipient with a smuggled header", email.Message{To: "user@example.org\r\nBcc: victim@example.org", Subject: "Hi", Text: "x"}},
		{"empty subject", email.Message{To: "user@example.org", Subject: " ", Text: "x"}},
		{"empty body", email.Message{To: "user@example.org", Subject: "Hi"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Port 1 is never listening: a validation failure must be reported before dialling.
			sender := email.NewSMTPSender(mustAddress(t, "no-reply@example.com"),
				email.SMTPConfig{Host: "127.0.0.1", Port: 1, TLS: email.SMTPTLSNone})

			err := sender.Send(t.Context(), tt.msg)

			assert.ErrorIs(t, err, email.ErrInvalidMessage)
		})
	}
}
