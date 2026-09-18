package config

import (
	"strings"
	"time"

	"goilerplate/pkg/filesystem"
)

type Config struct {
	App        App                `mapstructure:"app"`
	Server     Server             `mapstructure:"server"`
	GRPC       GRPC               `mapstructure:"grpc"`
	DB         DB                 `mapstructure:"db"`
	Redis      Redis              `mapstructure:"redis"`
	JWT        JWT                `mapstructure:"jwt"`
	Auth       Auth               `mapstructure:"auth"`
	Log        *Logger            `mapstructure:"log"`
	OTel       OTel               `mapstructure:"otel"`
	RateLimit  RateLimit          `mapstructure:"rate_limit"`
	FileSystem FileSystem         `mapstructure:"filesystem"`
	Crypto     Crypto             `mapstructure:"crypto"`
	Apikeys    map[string]string  `mapstructure:"api_key"`
	Services   map[string]Service `mapstructure:"service"`
}

type RateLimit struct {
	Auth    RateLimitRule `mapstructure:"auth"`
	User    RateLimitRule `mapstructure:"user"`
	Partner RateLimitRule `mapstructure:"partner"`
}

type RateLimitRule struct {
	Max        int           `mapstructure:"max"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type OTel struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"` // OTLP gRPC endpoint, e.g. "localhost:4317"
	Insecure bool   `mapstructure:"insecure"` // skip TLS — set true for local/dev
}

type GRPC struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type App struct {
	Env         string `mapstructure:"env"`
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Description string `mapstructure:"description"`
}

type Server struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Prefork      bool          `mapstructure:"prefork"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	EnableCORS   bool          `mapstructure:"enable_cors"`
	CORS         CORS
}

type CORS struct {
	AllowOrigin  string `mapstructure:"allow_origin"`
	AllowMethods string `mapstructure:"allow_methods"`
	AllowHeaders string `mapstructure:"allow_headers"`
}

type DB struct {
	Host                  string `mapstructure:"host"`
	Port                  int    `mapstructure:"port"`
	Name                  string `mapstructure:"name"`
	SSLMode               string `mapstructure:"sslmode"`
	Username              string `mapstructure:"username"`
	Password              string `mapstructure:"password"`
	MinOpenConnections    int    `mapstructure:"min_open_connections"`
	MaxOpenConnections    int    `mapstructure:"max_open_connections"`
	ConnectionMaxLifetime int    `mapstructure:"connection_max_lifetime"`
	ConnectionMaxIdleTime int    `mapstructure:"connection_max_idle_time"`
	HealthCheckPeriod     int    `mapstructure:"health_check_period"`
}

type Redis struct {
	Enabled      bool          `mapstructure:"enabled"`
	Host         string        `mapstructure:"host"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
}

type JWT struct {
	KeyID             string        `mapstructure:"key_id"`        // published as the "kid" header of every issued token
	AccessSecret      string        `mapstructure:"access_secret"` // active signing secret for access tokens
	RefreshSecret     string        `mapstructure:"refresh_secret"`
	PreviousKeys      []JWTKey      `mapstructure:"previous_keys"` // verification only, so secrets rotate without logging everyone out
	AccessTokenExpiry time.Duration `mapstructure:"access_token_expiry"`
	Issuer            string        `mapstructure:"issuer"`
	Audience          string        `mapstructure:"audience"`
	Leeway            time.Duration `mapstructure:"leeway"` // clock skew tolerance; jwt.DefaultLeeway when unset
}

// JWTKey is a retired signing key kept for verification during rotation.
type JWTKey struct {
	KeyID         string `mapstructure:"key_id"`
	AccessSecret  string `mapstructure:"access_secret"`
	RefreshSecret string `mapstructure:"refresh_secret"`
}

// Revocation modes for auth.revocation.
const (
	// RevocationStrict checks the session on every authenticated request, so logout and
	// admin deactivation take effect immediately (subject to the cache TTL).
	RevocationStrict = "strict"
	// RevocationRefreshOnly checks the session only when refreshing, so revocation takes
	// effect once the current access token expires. Cheaper, but not immediate.
	RevocationRefreshOnly = "refresh_only"
)

// Cache modes for auth.session_cache.
const (
	CacheModeAuto   = "auto"   // redis when redis.enabled, otherwise none
	CacheModeNone   = "none"   // no caching; every check reads the database
	CacheModeMemory = "memory" // in-process; revocation may lag across instances by session_cache_ttl
	CacheModeRedis  = "redis"  // shared by every instance
)

// Defaults applied when the corresponding auth settings are zero.
const (
	DefaultSessionCacheTTL    = 30 * time.Second
	DefaultPermissionCacheTTL = 15 * time.Minute
	DefaultSessionExpiry      = 7 * 24 * time.Hour
	DefaultRememberMeExpiry   = 30 * 24 * time.Hour
	DefaultRefreshReuseGrace  = 10 * time.Second
)

type Auth struct {
	SessionCache       string        `mapstructure:"session_cache"`        // auto | none | memory | redis (also used for the permission cache)
	SessionCacheTTL    time.Duration `mapstructure:"session_cache_ttl"`    // memory: max cross-instance revocation lag; redis: bounds staleness after direct DB edits
	PermissionCacheTTL time.Duration `mapstructure:"permission_cache_ttl"` // safety net; permission changes invalidate explicitly
	Revocation         string        `mapstructure:"revocation"`           // strict | refresh_only
	SessionExpiry      time.Duration `mapstructure:"session_expiry"`       // absolute session lifetime; also the refresh token lifetime
	RememberMeExpiry   time.Duration `mapstructure:"remember_me_expiry"`   // absolute session lifetime when remember_me = true
	RefreshReuseGrace  time.Duration `mapstructure:"refresh_reuse_grace"`  // window in which the just-replaced refresh token is still accepted
}

// RevocationMode resolves auth.revocation, defaulting to strict (secure by default).
func (a Auth) RevocationMode() string {
	mode := strings.ToLower(strings.TrimSpace(a.Revocation))
	if mode == RevocationRefreshOnly {
		return RevocationRefreshOnly
	}
	return RevocationStrict
}

// CacheMode resolves auth.session_cache, turning "auto" (or empty) into redis or none.
func (a Auth) CacheMode(redisEnabled bool) string {
	mode := strings.ToLower(strings.TrimSpace(a.SessionCache))
	if mode != "" && mode != CacheModeAuto {
		return mode
	}
	if redisEnabled {
		return CacheModeRedis
	}
	return CacheModeNone
}

// SessionCacheTTLOrDefault returns session_cache_ttl, or DefaultSessionCacheTTL when unset.
func (a Auth) SessionCacheTTLOrDefault() time.Duration {
	if a.SessionCacheTTL > 0 {
		return a.SessionCacheTTL
	}
	return DefaultSessionCacheTTL
}

// PermissionCacheTTLOrDefault returns permission_cache_ttl, or DefaultPermissionCacheTTL when unset.
func (a Auth) PermissionCacheTTLOrDefault() time.Duration {
	if a.PermissionCacheTTL > 0 {
		return a.PermissionCacheTTL
	}
	return DefaultPermissionCacheTTL
}

// SessionExpiryOrDefault returns session_expiry, or DefaultSessionExpiry when unset.
func (a Auth) SessionExpiryOrDefault() time.Duration {
	if a.SessionExpiry > 0 {
		return a.SessionExpiry
	}
	return DefaultSessionExpiry
}

// RememberMeExpiryOrDefault returns remember_me_expiry, or DefaultRememberMeExpiry when unset.
func (a Auth) RememberMeExpiryOrDefault() time.Duration {
	if a.RememberMeExpiry > 0 {
		return a.RememberMeExpiry
	}
	return DefaultRememberMeExpiry
}

// RefreshReuseGraceOrDefault returns refresh_reuse_grace, or DefaultRefreshReuseGrace when
// unset. Zero is not a valid override: it would make concurrent tabs log each other out.
func (a Auth) RefreshReuseGraceOrDefault() time.Duration {
	if a.RefreshReuseGrace > 0 {
		return a.RefreshReuseGrace
	}
	return DefaultRefreshReuseGrace
}

type Logger struct {
	Level         string   `mapstructure:"level"`
	Source        bool     `mapstructure:"source"`
	RedactFields  []string `mapstructure:"redact_fields"`   // extra field names masked in logs, on top of pkg/redact defaults
	OmitBodyPaths []string `mapstructure:"omit_body_paths"` // path prefixes whose request/response bodies are never logged
}

type FileSystem struct {
	Driver      string                 `mapstructure:"driver"`        // local, s3, drive
	MaxFileSize int64                  `mapstructure:"max_file_size"` // Maximum file size in bytes
	Local       filesystem.LocalConfig `mapstructure:"local"`
	S3          filesystem.S3Config    `mapstructure:"s3"`
	Drive       filesystem.DriveConfig `mapstructure:"drive"`
}

type Crypto struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

type Service struct {
	Name    string `mapstructure:"name"`
	BaseURL string `mapstructure:"base_url"`
	Apikey  string `mapstructure:"api_key"`
}
