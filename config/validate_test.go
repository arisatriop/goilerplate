package config

import (
	"strings"
	"testing"
	"time"

	"goilerplate/pkg/filesystem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validAccessSecret  = "k3Jv9QpX2mZr7TbW5nLc8HsYd4FgA1uE"
	validRefreshSecret = "Pz6Rw1YtN8qLm3XcV7bK2jHf5Gd9Sa4U"
)

func validConfig() *Config {
	return &Config{
		App:    App{Env: "dev", Name: "goilerplate"},
		Server: Server{Port: 3000},
		DB: DB{
			Driver: "postgres", Host: "localhost", Port: 5432,
			Name: "goilerplate", Username: "postgres", Password: "postgres", MaxOpenConnections: 10,
		},
		JWT: JWT{
			SecretKey:          "Lg7-legacy-secret-Q2x",
			AccessSecret:       validAccessSecret,
			RefreshSecret:      validRefreshSecret,
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
			Issuer:             "goilerplate",
		},
		FileSystem: FileSystem{Driver: "local", MaxFileSize: 1024, Local: filesystem.LocalConfig{BasePath: "./storage"}},
		Apikeys:    map[string]string{"default": "9f1c2e7a-4b3d-4a8e-9c61-2d5f7b8e0a13"},
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := validConfig()

	assert.NoError(t, cfg.Validate())
}

func TestConfig_Validate_Rules(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{"redis enabled without host", func(c *Config) { c.Redis.Enabled = true }, "redis.host is required"},
		{"redis disabled ignores host", func(c *Config) { c.Redis.Host = "" }, ""},
		{"redis host placeholder", func(c *Config) { c.Redis = Redis{Enabled: true, Host: "<REDIS_HOST>"} }, "redis.host still contains the placeholder"},
		{"s3 without bucket", func(c *Config) { c.FileSystem.Driver = "s3"; c.FileSystem.S3.Region = "us-east-1" }, "filesystem.s3.bucket is required"},
		{"s3 without region", func(c *Config) { c.FileSystem.Driver = "s3"; c.FileSystem.S3.Bucket = "b" }, "filesystem.s3.region is required"},
		{"unknown filesystem driver", func(c *Config) { c.FileSystem.Driver = "ftp" }, `filesystem.driver must be local, s3, or drive, got "ftp"`},
		{"drive without credentials", func(c *Config) { c.FileSystem.Driver = "drive" }, "filesystem.drive requires credentials_file"},
		{"grpc enabled without port", func(c *Config) { c.GRPC.Enabled = true }, "grpc.port must be between 1 and 65535, got 0"},
		{"grpc port equals server port", func(c *Config) { c.GRPC = GRPC{Enabled: true, Port: 3000} }, "grpc.port must differ from server.port"},
		{"grpc disabled ignores port", func(c *Config) { c.GRPC.Port = 0 }, ""},
		{"otel enabled without endpoint", func(c *Config) { c.OTel.Enabled = true }, "otel.endpoint is required"},
		{"db placeholder", func(c *Config) { c.DB.Host = "<DB_HOST>" }, "db.host still contains the placeholder <DB_HOST>"},
		{"db unsupported driver", func(c *Config) { c.DB.Driver = "sqlite" }, `db.driver must be postgres or mysql, got "sqlite"`},
		{"access secret too short", func(c *Config) { c.JWT.AccessSecret = "short" }, "jwt.access_secret must be at least 32 bytes, got 5"},
		{"refresh secret missing", func(c *Config) { c.JWT.RefreshSecret = "" }, "jwt.refresh_secret is required"},
		{"secret placeholder", func(c *Config) { c.JWT.AccessSecret = "<JWT_ACCESS_SECRET_KEY>" }, "jwt.access_secret still contains a placeholder"},
		{"legacy secret placeholder", func(c *Config) { c.JWT.SecretKey = "<JWT_SECRET_KEY>" }, "jwt.secret_key still contains a placeholder"},
		{"secrets identical", func(c *Config) { c.JWT.RefreshSecret = validAccessSecret }, "jwt.access_secret and jwt.refresh_secret must be different"},
		{"access expiry zero", func(c *Config) { c.JWT.AccessTokenExpiry = 0 }, "jwt.access_token_expiry must be greater than 0"},
		{"refresh expiry not longer", func(c *Config) { c.JWT.RefreshTokenExpiry = time.Minute }, "jwt.refresh_token_expiry must be greater than jwt.access_token_expiry"},
		{"api key placeholder", func(c *Config) { c.Apikeys["partner1"] = "<API_KEY_PARTNER1>" }, "api_key.partner1 still contains the placeholder"},
		{"db pool size zero", func(c *Config) { c.DB.MaxOpenConnections = 0 }, "db.max_open_connections must be at least 1, got 0"},
		{"unknown session cache", func(c *Config) { c.Auth.SessionCache = "memcached" }, `auth.session_cache must be auto, none, memory, or redis, got "memcached"`},
		{"redis session cache without redis", func(c *Config) { c.Auth.SessionCache = "redis" }, "auth.session_cache=redis requires redis.enabled=true"},
		{"memory session cache", func(c *Config) { c.Auth.SessionCache = "memory" }, ""},
		{"negative session cache ttl", func(c *Config) { c.Auth.SessionCacheTTL = -time.Second }, "auth.session_cache_ttl must not be negative"},
		{"negative permission cache ttl", func(c *Config) { c.Auth.PermissionCacheTTL = -time.Second }, "auth.permission_cache_ttl must not be negative"},
		{"max file size zero", func(c *Config) { c.FileSystem.MaxFileSize = 0 }, "filesystem.max_file_size must be greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := validConfig()
			tt.mutate(cfg)

			// Act
			err := cfg.Validate()

			// Assert
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfig_Validate_ProductionRejectsWeakSecrets(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{"example marker", func(c *Config) { c.JWT.AccessSecret = "please-changeme-access-secret-value!!" }, `jwt.access_secret looks like an example value (contains "changeme")`},
		{"low variety", func(c *Config) { c.JWT.RefreshSecret = strings.Repeat("ab", 20) }, "jwt.refresh_secret has too little variety (2 distinct characters)"},
		{"sample api key", func(c *Config) { c.Apikeys["default"] = "your-api-key-here-0123456789" }, `api_key.default looks like an example value`},
		{"sample s3 secret", func(c *Config) {
			c.FileSystem = FileSystem{Driver: "s3", MaxFileSize: 1, S3: filesystem.S3Config{Bucket: "b", Region: "r", SecretAccessKey: "YOUR_AWS_SECRET_ACCESS_KEY"}}
		}, "filesystem.s3.secret_access_key looks like an example value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := validConfig()
			cfg.App.Env = "Production"
			tt.mutate(cfg)

			// Act
			err := cfg.Validate()

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfig_Validate_NonProductionAllowsWeakSecrets(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.AccessSecret = "local-dev-changeme-access-secret-000"

	assert.NoError(t, cfg.Validate())
}

func TestConfig_Validate_ReportsAllErrorsWithoutSecrets(t *testing.T) {
	// Arrange
	cfg := validConfig()
	cfg.App.Env = "production"
	cfg.Redis.Enabled = true
	cfg.FileSystem.Local.BasePath = ""
	cfg.JWT.AccessSecret = "tooshort-secret"
	cfg.JWT.RefreshSecret = "tooshort-secret"

	// Act
	err := cfg.Validate()

	// Assert
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "redis.host is required")
	assert.Contains(t, msg, "filesystem.local.base_path is required")
	assert.Contains(t, msg, "jwt.access_secret must be at least 32 bytes")
	assert.Contains(t, msg, "jwt.refresh_secret must be at least 32 bytes")
	assert.Contains(t, msg, "must be different")
	assert.NotContains(t, msg, "tooshort-secret")
	assert.NotContains(t, msg, "Lg7-legacy-secret-Q2x")
}

func TestAuth_CacheMode(t *testing.T) {
	tests := []struct {
		configured   string
		redisEnabled bool
		want         string
	}{
		{"", false, CacheModeNone},
		{"", true, CacheModeRedis},
		{"auto", false, CacheModeNone},
		{" AUTO ", true, CacheModeRedis},
		{"none", true, CacheModeNone},
		{"memory", true, CacheModeMemory},
		{"redis", true, CacheModeRedis},
	}

	for _, tt := range tests {
		got := Auth{SessionCache: tt.configured}.CacheMode(tt.redisEnabled)
		assert.Equal(t, tt.want, got, "session_cache=%q redis=%v", tt.configured, tt.redisEnabled)
	}
}

func TestAuth_TTLDefaults(t *testing.T) {
	assert.Equal(t, DefaultSessionCacheTTL, Auth{}.SessionCacheTTLOrDefault())
	assert.Equal(t, DefaultPermissionCacheTTL, Auth{}.PermissionCacheTTLOrDefault())

	custom := Auth{SessionCacheTTL: 5 * time.Second, PermissionCacheTTL: time.Minute}
	assert.Equal(t, 5*time.Second, custom.SessionCacheTTLOrDefault())
	assert.Equal(t, time.Minute, custom.PermissionCacheTTLOrDefault())
}
