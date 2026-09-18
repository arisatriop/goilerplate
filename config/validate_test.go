package config

import (
	"strings"
	"testing"
	"time"

	"goilerplate/pkg/apikey"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/hash"

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
			Host: "localhost", Port: 5432,
			Name: "goilerplate", Username: "postgres", Password: "postgres", MaxOpenConnections: 10,
		},
		JWT: JWT{
			KeyID:             "v1",
			AccessSecret:      validAccessSecret,
			RefreshSecret:     validRefreshSecret,
			AccessTokenExpiry: 15 * time.Minute,
			Issuer:            "goilerplate",
			Audience:          "goilerplate-api",
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
		{"access secret too short", func(c *Config) { c.JWT.AccessSecret = "short" }, "jwt.access_secret must be at least 32 bytes, got 5"},
		{"refresh secret missing", func(c *Config) { c.JWT.RefreshSecret = "" }, "jwt.refresh_secret is required"},
		{"secret placeholder", func(c *Config) { c.JWT.AccessSecret = "<JWT_ACCESS_SECRET_KEY>" }, "jwt.access_secret still contains a placeholder"},
		{"key id missing", func(c *Config) { c.JWT.KeyID = "" }, "jwt.key_id is required"},
		{"audience missing", func(c *Config) { c.JWT.Audience = "" }, "jwt.audience is required"},
		{"secrets identical", func(c *Config) { c.JWT.RefreshSecret = validAccessSecret }, "jwt.access_secret and jwt.refresh_secret must be different"},
		{"access expiry zero", func(c *Config) { c.JWT.AccessTokenExpiry = 0 }, "jwt.access_token_expiry must be greater than 0"},
		{"negative leeway", func(c *Config) { c.JWT.Leeway = -time.Second }, "jwt.leeway must not be negative"},
		{"previous key without id", func(c *Config) {
			c.JWT.PreviousKeys = []JWTKey{{AccessSecret: validAccessSecret, RefreshSecret: validRefreshSecret}}
		}, "jwt.previous_keys[0].key_id is required"},
		{"previous key reuses active id", func(c *Config) {
			c.JWT.PreviousKeys = []JWTKey{{KeyID: "v1", AccessSecret: validAccessSecret, RefreshSecret: validRefreshSecret}}
		}, `jwt.previous_keys[0].key_id "v1" is already used by another key`},
		{"previous key secret too short", func(c *Config) {
			c.JWT.PreviousKeys = []JWTKey{{KeyID: "v0", AccessSecret: "short", RefreshSecret: validRefreshSecret}}
		}, "jwt.previous_keys[0].access_secret must be at least 32 bytes, got 5"},
		{"valid previous key", func(c *Config) {
			c.JWT.PreviousKeys = []JWTKey{{KeyID: "v0", AccessSecret: validAccessSecret, RefreshSecret: validRefreshSecret}}
		}, ""},
		{"session expiry shorter than access expiry", func(c *Config) { c.Auth.SessionExpiry = time.Minute }, "auth.session_expiry must be greater than jwt.access_token_expiry"},
		{"negative session expiry", func(c *Config) { c.Auth.SessionExpiry = -time.Second }, "auth.session_expiry must not be negative"},
		{"remember me shorter than session", func(c *Config) {
			c.Auth.SessionExpiry = 48 * time.Hour
			c.Auth.RememberMeExpiry = time.Hour
		}, "auth.remember_me_expiry must not be shorter than auth.session_expiry"},
		{"api key placeholder", func(c *Config) { c.Apikeys["partner1"] = "<API_KEY_PARTNER1>" }, "api_key.partner1 still contains the placeholder"},
		{"api key digest too short", func(c *Config) {
			c.Apikeys["partner1"] = apikey.HashPrefix + strings.Repeat("a", 63)
		}, "api_key.partner1 must be sha256: followed by 64 hex characters, got 63"},
		{"api key digest not hex", func(c *Config) {
			c.Apikeys["partner1"] = apikey.HashPrefix + strings.Repeat("z", 64)
		}, "api_key.partner1 must be sha256: followed by hex characters"},
		{"api key digest", func(c *Config) {
			c.Apikeys["partner1"] = apikey.HashPrefix + strings.Repeat("ab", 32)
		}, ""},
		{"db pool size zero", func(c *Config) { c.DB.MaxOpenConnections = 0 }, "db.max_open_connections must be at least 1, got 0"},
		{"bad trusted proxy CIDR", func(c *Config) { c.Server.TrustedProxies = []string{"10.0.0.0/99"} }, `server.trusted_proxies[0] "10.0.0.0/99" is not a valid CIDR`},
		{"bad trusted proxy IP", func(c *Config) { c.Server.TrustedProxies = []string{"not-an-ip"} }, `server.trusted_proxies[0] "not-an-ip" is not a valid IP or CIDR`},
		{"empty trusted proxy", func(c *Config) { c.Server.TrustedProxies = []string{"  "} }, "server.trusted_proxies[0] must not be empty"},
		{"valid trusted proxies", func(c *Config) { c.Server.TrustedProxies = []string{"10.0.0.0/8", "192.0.2.1"} }, ""},
		{"unknown revocation mode", func(c *Config) { c.Auth.Revocation = "sometimes" }, `auth.revocation must be strict or refresh_only, got "sometimes"`},
		{"refresh_only revocation", func(c *Config) { c.Auth.Revocation = "refresh_only" }, ""},
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

func TestAuth_RevocationMode(t *testing.T) {
	// Unset and unknown values fall back to strict: revocation must be secure by default.
	assert.Equal(t, RevocationStrict, Auth{}.RevocationMode())
	assert.Equal(t, RevocationStrict, Auth{Revocation: "strict"}.RevocationMode())
	assert.Equal(t, RevocationStrict, Auth{Revocation: "  STRICT  "}.RevocationMode())
	assert.Equal(t, RevocationRefreshOnly, Auth{Revocation: "refresh_only"}.RevocationMode())
	assert.Equal(t, RevocationRefreshOnly, Auth{Revocation: " Refresh_Only "}.RevocationMode())
}

func TestServer_ProxyHeaderOrDefault(t *testing.T) {
	assert.Equal(t, "X-Forwarded-For", Server{}.ProxyHeaderOrDefault())
	assert.Equal(t, "X-Forwarded-For", Server{ProxyHeader: "  "}.ProxyHeaderOrDefault())
	assert.Equal(t, "CF-Connecting-IP", Server{ProxyHeader: "CF-Connecting-IP"}.ProxyHeaderOrDefault())
}

// A digest of a weak key looks exactly as strong as a digest of a random one, so the entropy
// heuristics must not be applied to it — claiming to have checked something SHA-256 hides would
// be worse than not checking. The plaintext form is still checked.
func TestConfig_Validate_ProductionSkipsEntropyChecksOnHashedAPIKeys(t *testing.T) {
	weak := "changeme"

	hashed := validConfig()
	hashed.App.Env = "production"
	hashed.Apikeys["default"] = apikey.HashPrefix + hash.Token(weak)
	assert.NoError(t, hashed.Validate(), "a digest is accepted on its shape alone")

	plain := validConfig()
	plain.App.Env = "production"
	plain.Apikeys["default"] = weak
	require.Error(t, plain.Validate())
	assert.Contains(t, plain.Validate().Error(), `api_key.default looks like an example value`)
}
