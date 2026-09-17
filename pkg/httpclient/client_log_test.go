package httpclient

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goilerplate/pkg/redact"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingRoundTripper_RedactsSecrets(t *testing.T) {
	// Arrange
	t.Setenv("APP_ENV", "test")
	logs := &bytes.Buffer{}
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "session=cookie-secret")
		_, _ = w.Write([]byte(`{"message":"ok","access_token":"upstream-token"}`))
	}))
	t.Cleanup(server.Close)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/charge?api_key=query-secret&ref=42",
		strings.NewReader(`{"client_secret":"body-secret","amount":"10.00"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic header-secret")

	// Act
	resp, err := NewClient(0).Do(req)

	// Assert
	require.NoError(t, err)
	_ = resp.Body.Close()
	output := logs.String()
	assert.Contains(t, output, LogLabel)
	for _, secret := range []string{"cookie-secret", "upstream-token", "query-secret", "body-secret", "header-secret"} {
		assert.NotContains(t, output, secret)
	}
	assert.Contains(t, output, redact.Mask)
	assert.Contains(t, output, "ref=42")
}
