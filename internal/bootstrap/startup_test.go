package bootstrap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"goilerplate/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An unreachable Redis used to os.Exit from inside bootstrap, skipping every deferred cleanup and
// taking the decision away from the caller. It must come back as an error instead.
func TestNewRedis_UnreachableIsAnError(t *testing.T) {
	cfg := &config.Config{Redis: config.Redis{
		Enabled:     true,
		Host:        "127.0.0.1:1", // nothing listens on port 1
		DialTimeout: 200 * time.Millisecond,
	}}

	client, err := NewRedis(cfg)

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "connecting to redis")
}

func TestNewRedis_DisabledIsNil(t *testing.T) {
	client, err := NewRedis(&config.Config{})

	require.NoError(t, err)
	assert.Nil(t, client)
}

func captureDefaultLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &buf
}

// Config validation reports every problem at once; the startup log must keep them as separate
// entries so an operator can read them, rather than one string with embedded newlines.
func TestLogStartupFailure_ListsEachConfigurationProblem(t *testing.T) {
	logs := captureDefaultLog(t)
	err := fmt.Errorf("invalid configuration: %w",
		errors.Join(errors.New("db.host is required"), errors.New("jwt.key_id is required")))

	LogStartupFailure(err)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "startup failed", entry["msg"])
	assert.Equal(t, "invalid configuration", entry["error"], "the problems are listed once, not repeated in error")
	assert.Equal(t, []any{"db.host is required", "jwt.key_id is required"}, entry["problems"])
}

func TestLogStartupFailure_SingleError(t *testing.T) {
	logs := captureDefaultLog(t)

	LogStartupFailure(errors.New("connecting to postgres: refused"))

	var entry map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	assert.Equal(t, "connecting to postgres: refused", entry["error"])
	assert.NotContains(t, entry, "problems")
}
