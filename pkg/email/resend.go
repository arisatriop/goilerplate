package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

// DefaultResendBaseURL is Resend's API endpoint.
const DefaultResendBaseURL = "https://api.resend.com"

// ResendConfig configures the resend driver.
type ResendConfig struct {
	// APIKey is a Resend API key with sending access. Set it from the environment
	// (EMAIL_RESEND_API_KEY), never in a committed file.
	APIKey string `mapstructure:"api_key"`
	// BaseURL overrides the API endpoint. Empty means DefaultResendBaseURL; only a test or an
	// egress proxy needs another.
	BaseURL string `mapstructure:"base_url"`
	// Timeout bounds one request. Default 30s.
	Timeout time.Duration `mapstructure:"timeout"`
}

// ResendSender delivers through the Resend HTTP API. It needs no SDK: the API is one POST.
type ResendSender struct {
	from    *mail.Address
	cfg     ResendConfig
	baseURL string
	client  *http.Client
}

// NewResendSender returns a sender for cfg.
func NewResendSender(from *mail.Address, cfg ResendConfig) *ResendSender {
	baseURL := strings.TrimSuffix(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultResendBaseURL
	}

	return &ResendSender{
		from:    from,
		cfg:     cfg,
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeoutOrDefault(cfg.Timeout)},
	}
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html,omitempty"`
}

// maxErrorBody caps how much of a failed response is kept in the error. Provider errors are
// short; anything longer is not worth carrying into a log line.
const maxErrorBody = 512

// Send delivers msg.
func (s *ResendSender) Send(ctx context.Context, msg Message) error {
	to, err := msg.validate()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(resendRequest{
		From:    formatAddress(s.from),
		To:      []string{formatAddress(to)},
		Subject: msg.Subject,
		Text:    msg.Text,
		HTML:    msg.HTML,
	})
	if err != nil {
		return fmt.Errorf("resend: encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("resend: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody)) // best effort: the status alone is enough
		return fmt.Errorf("resend: status %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}

	// Drain so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
