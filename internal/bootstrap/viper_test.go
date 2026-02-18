package bootstrap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	cfg := Load()

	assert.Equal(t, "TestApp", cfg.App.Name)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, "supersecret", cfg.JWT.AccessSecret)
	assert.True(t, cfg.Redis.Enabled)
	assert.Equal(t, "api-key-default", cfg.Apikeys["default"])
	assert.Equal(t, "GOPAY-TEST", cfg.Services["gopay"].Name)
	assert.Equal(t, "https://test.gopay.co.id", cfg.Services["gopay"].BaseURL)
}
