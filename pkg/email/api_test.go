package email_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goilerplate/pkg/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturedRequest is what a fake provider endpoint received.
type capturedRequest struct {
	method string
	path   string
	header http.Header
	body   map[string]any
}

func newFakeProvider(t *testing.T, status int, reply string) (*httptest.Server, *capturedRequest) {
	t.Helper()

	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		captured.method, captured.path, captured.header = r.Method, r.URL.Path, r.Header.Clone()
		assert.NoError(t, json.Unmarshal(raw, &captured.body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(server.Close)
	return server, captured
}

var testMessage = email.Message{
	To:      "user@example.org",
	Subject: "Reset your password",
	Text:    "Open the link",
	HTML:    "<p>Open the link</p>",
}

func TestResendSender_PostsTheMessage(t *testing.T) {
	server, captured := newFakeProvider(t, http.StatusOK, `{"id":"abc"}`)
	sender := email.NewResendSender(mustAddress(t, "Goilerplate <no-reply@example.com>"),
		email.ResendConfig{APIKey: "re_test_key", BaseURL: server.URL})

	require.NoError(t, sender.Send(t.Context(), testMessage))

	assert.Equal(t, http.MethodPost, captured.method)
	assert.Equal(t, "/emails", captured.path)
	assert.Equal(t, "Bearer re_test_key", captured.header.Get("Authorization"))
	assert.Equal(t, `"Goilerplate" <no-reply@example.com>`, captured.body["from"])
	assert.Equal(t, []any{"user@example.org"}, captured.body["to"])
	assert.Equal(t, "Reset your password", captured.body["subject"])
	assert.Equal(t, "Open the link", captured.body["text"])
	assert.Equal(t, "<p>Open the link</p>", captured.body["html"])
}

func TestResendSender_ReportsARejection(t *testing.T) {
	server, _ := newFakeProvider(t, http.StatusUnprocessableEntity,
		`{"name":"validation_error","message":"The from domain is not verified"}`)
	sender := email.NewResendSender(mustAddress(t, "no-reply@example.com"),
		email.ResendConfig{APIKey: "re_test_key", BaseURL: server.URL})

	err := sender.Send(t.Context(), testMessage)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "422")
	assert.Contains(t, err.Error(), "not verified", "the provider's reason is kept for the log")
}

func TestSESSender_SendsThroughTheV2API(t *testing.T) {
	server, captured := newFakeProvider(t, http.StatusOK, `{"MessageId":"abc"}`)
	sender, err := email.NewSESSender(t.Context(), mustAddress(t, "no-reply@example.com"), email.SESConfig{
		Region:           "ap-southeast-1",
		AccessKeyID:      "AKIDEXAMPLE",
		SecretAccessKey:  "not-a-real-secret",
		ConfigurationSet: "transactional",
		Endpoint:         server.URL,
	})
	require.NoError(t, err)

	require.NoError(t, sender.Send(t.Context(), testMessage))

	assert.Equal(t, http.MethodPost, captured.method)
	assert.Equal(t, "/v2/email/outbound-emails", captured.path)
	assert.True(t, strings.HasPrefix(captured.header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/"),
		"the request is signed with the configured credentials")
	assert.Equal(t, "no-reply@example.com", captured.body["FromEmailAddress"])
	assert.Equal(t, "transactional", captured.body["ConfigurationSetName"])

	simple := captured.body["Content"].(map[string]any)["Simple"].(map[string]any)
	assert.Equal(t, "Reset your password", simple["Subject"].(map[string]any)["Data"])
	assert.Equal(t, "Open the link", simple["Body"].(map[string]any)["Text"].(map[string]any)["Data"])
	assert.Equal(t, "<p>Open the link</p>", simple["Body"].(map[string]any)["Html"].(map[string]any)["Data"])
}

func TestNew_SelectsTheConfiguredDriver(t *testing.T) {
	tests := []struct {
		name    string
		cfg     email.Config
		want    any
		wantErr string
	}{
		{"unset is log", email.Config{From: "no-reply@example.com"}, &email.LogSender{}, ""},
		{"smtp", email.Config{Driver: "SMTP", From: "no-reply@example.com"}, &email.SMTPSender{}, ""},
		{"resend", email.Config{Driver: "resend", From: "no-reply@example.com"}, &email.ResendSender{}, ""},
		{"ses", email.Config{Driver: "ses", From: "no-reply@example.com", SES: email.SESConfig{Region: "us-east-1"}}, &email.SESSender{}, ""},
		{"unknown driver", email.Config{Driver: "pigeon", From: "no-reply@example.com"}, nil, `unknown email driver "pigeon"`},
		{"malformed from", email.Config{From: "nobody"}, nil, "email.from"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := email.New(t.Context(), tt.cfg, nil)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.IsType(t, tt.want, sender)
		})
	}
}
