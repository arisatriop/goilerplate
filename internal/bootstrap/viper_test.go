package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"goilerplate/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_PureEnv(t *testing.T) {
	os.Setenv("APP_NAME", "TestApp")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("JWT_ACCESS_SECRET", "supersecret")
	os.Setenv("REDIS_ENABLED", "true")
	os.Setenv("API_KEY_DEFAULT", "api-key-default")
	os.Setenv("SERVICE_GOPAY_NAME", "GOPAY-TEST")
	os.Setenv("SERVICE_GOPAY_BASE_URL", "https://test.gopay.co.id")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "TestApp", cfg.App.Name)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, "supersecret", cfg.JWT.AccessSecret)
	assert.True(t, cfg.Redis.Enabled)
	assert.Equal(t, "api-key-default", cfg.Apikeys["default"])
	assert.Equal(t, "GOPAY-TEST", cfg.Services["gopay"].Name)
	assert.Equal(t, "https://test.gopay.co.id", cfg.Services["gopay"].BaseURL)
}

// loadWithConfigFile runs Load from a directory holding the given config.yaml, and with the
// previous-key variables set only as the test asks.
func loadWithConfigFile(t *testing.T, yaml string, env map[string]string) *config.Config {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config", "config.yaml"), []byte(yaml), 0o600))
	t.Chdir(dir)

	for _, name := range []string{"JWT_PREVIOUS_KEY_ID", "JWT_PREVIOUS_ACCESS_SECRET", "JWT_PREVIOUS_REFRESH_SECRET"} {
		t.Setenv(name, env[name])
	}

	cfg, err := Load()
	require.NoError(t, err)
	return cfg
}

func TestLoad_PreviousJWTKeyFromEnv(t *testing.T) {
	const (
		access  = "previous-access-value-0123456789-abcdef"
		refresh = "previous-refresh-value-0123456789-abcdef"
	)
	fromEnv := config.JWTKey{KeyID: "v1", AccessSecret: access, RefreshSecret: refresh}
	allThree := map[string]string{
		"JWT_PREVIOUS_KEY_ID":         "v1",
		"JWT_PREVIOUS_ACCESS_SECRET":  access,
		"JWT_PREVIOUS_REFRESH_SECRET": refresh,
	}
	fromFile := config.JWTKey{KeyID: "v0", AccessSecret: "file-access", RefreshSecret: "file-refresh"}
	withFileKey := "jwt:\n  previous_keys:\n    - key_id: v0\n      access_secret: file-access\n      refresh_secret: file-refresh\n"

	tests := []struct {
		name string
		yaml string
		env  map[string]string
		want []config.JWTKey
	}{
		{name: "no variables adds nothing", yaml: "jwt:\n  key_id: v2\n"},
		{name: "all three variables add the key", yaml: "jwt:\n  key_id: v2\n", env: allThree, want: []config.JWTKey{fromEnv}},
		{name: "a key in config.yaml is kept", yaml: withFileKey, env: allThree, want: []config.JWTKey{fromFile, fromEnv}},
		{name: "no variables leaves config.yaml alone", yaml: withFileKey, want: []config.JWTKey{fromFile}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadWithConfigFile(t, tt.yaml, tt.env)

			assert.Equal(t, tt.want, cfg.JWT.PreviousKeys)
		})
	}
}

func TestLoad_PreviousJWTKeyFromEnv_InvalidIsRefusedAtStartup(t *testing.T) {
	valid := "previous-secret-value-0123456789-abcdef"

	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "missing secret",
			env:  map[string]string{"JWT_PREVIOUS_KEY_ID": "v1", "JWT_PREVIOUS_ACCESS_SECRET": valid},
			want: "jwt.previous_keys[0].refresh_secret is required",
		},
		{
			name: "missing key id",
			env:  map[string]string{"JWT_PREVIOUS_ACCESS_SECRET": valid, "JWT_PREVIOUS_REFRESH_SECRET": valid},
			want: "jwt.previous_keys[0].key_id",
		},
		{
			name: "key id of the active key",
			env:  map[string]string{"JWT_PREVIOUS_KEY_ID": "v2", "JWT_PREVIOUS_ACCESS_SECRET": valid, "JWT_PREVIOUS_REFRESH_SECRET": valid},
			want: "is already used by another key",
		},
		{
			name: "secret too short",
			env:  map[string]string{"JWT_PREVIOUS_KEY_ID": "v1", "JWT_PREVIOUS_ACCESS_SECRET": "short", "JWT_PREVIOUS_REFRESH_SECRET": valid},
			want: "jwt.previous_keys[0].access_secret must be at least 32 bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadWithConfigFile(t, "jwt:\n  key_id: v2\n", tt.env)

			err := cfg.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
