package config

import (
	"strings"
	"time"

	"goilerplate/pkg/filesystem"
)

type Config struct {
	App        App               `mapstructure:"app"`
	Server     Server            `mapstructure:"server"`
	GRPC       GRPC              `mapstructure:"grpc"`
	DB         DB                `mapstructure:"db"`
	Redis      Redis             `mapstructure:"redis"`
	JWT        JWT               `mapstructure:"jwt"`
	Auth       Auth              `mapstructure:"auth"`
	Log        *Logger           `mapstructure:"log"`
	OTel       OTel              `mapstructure:"otel"`
	RateLimit  RateLimit         `mapstructure:"rate_limit"`
	FileSystem FileSystem        `mapstructure:"filesystem"`
	Crypto     Crypto            `mapstructure:"crypto"`
	Apikeys    map[string]string `mapstructure:"api_key"`
	// Services is an extension point for upstreams this application calls; see Service.
	Services     map[string]Service `mapstructure:"service"`
	InternalAuth InternalAuth       `mapstructure:"internal_auth"`
	Jobs         Jobs               `mapstructure:"jobs"`
}

// Jobs holds the background jobs the application runs itself.
type Jobs struct {
	Cleanup Cleanup `mapstructure:"cleanup"`
}

// Defaults applied when the corresponding cleanup settings are zero.
const (
	DefaultCleanupInterval  = time.Hour
	DefaultCleanupRetention = 30 * 24 * time.Hour
	DefaultCleanupBatchSize = 1000
)

// Cleanup removes rows that can no longer be used: revoked or expired sessions, and one-time
// tokens that were consumed or have expired. They are kept for Retention past that moment so
// they can still answer "what happened to this login" during an incident.
//
// Off by default. Deleting rows is not something a boilerplate should start doing to a
// deployment that never asked for it.
type Cleanup struct {
	Enabled   bool          `mapstructure:"enabled"`
	Interval  time.Duration `mapstructure:"interval"`
	Retention time.Duration `mapstructure:"retention"`
	// BatchSize bounds one DELETE, so a first run against a large table does not hold locks
	// or bloat WAL for the length of the whole backlog.
	BatchSize int `mapstructure:"batch_size"`
}

// IntervalOrDefault returns jobs.cleanup.interval, or DefaultCleanupInterval when unset.
func (c Cleanup) IntervalOrDefault() time.Duration {
	if c.Interval > 0 {
		return c.Interval
	}
	return DefaultCleanupInterval
}

// RetentionOrDefault returns jobs.cleanup.retention, or DefaultCleanupRetention when unset.
func (c Cleanup) RetentionOrDefault() time.Duration {
	if c.Retention > 0 {
		return c.Retention
	}
	return DefaultCleanupRetention
}

// BatchSizeOrDefault returns jobs.cleanup.batch_size, or DefaultCleanupBatchSize when unset.
func (c Cleanup) BatchSizeOrDefault() int {
	if c.BatchSize > 0 {
		return c.BatchSize
	}
	return DefaultCleanupBatchSize
}

// Internal auth modes. See InternalAuth.
const (
	InternalAuthNone         = "none"
	InternalAuthSharedSecret = "shared_secret"
)

// InternalAuth guards the /internal routes, which are meant for pod-to-pod traffic only (D5).
//
// Keeping them off the public internet is the gateway's job: it forwards an explicit allowlist
// and never a catch-all. `shared_secret` is a second lock for when that gateway config is owned
// by another team, or changes often enough that one day it will be wrong.
type InternalAuth struct {
	// Mode is none or shared_secret. Default none.
	Mode string `mapstructure:"mode"`
	// Secret is compared against the X-Internal-Secret header. Required in shared_secret mode.
	Secret string `mapstructure:"secret"`
}

// ModeOrDefault returns internal_auth.mode, or InternalAuthNone when unset.
func (i InternalAuth) ModeOrDefault() string {
	mode := strings.ToLower(strings.TrimSpace(i.Mode))
	if mode == "" {
		return InternalAuthNone
	}
	return mode
}

// RequiresSecret reports whether callers must present X-Internal-Secret.
func (i InternalAuth) RequiresSecret() bool {
	return i.ModeOrDefault() == InternalAuthSharedSecret
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
	// Reflection lets a client discover services without the proto module. Off by default:
	// it publishes the full service surface to anyone who can reach the port.
	Reflection bool     `mapstructure:"reflection"`
	TLS        GRPCTLS  `mapstructure:"tls"`
	Auth       GRPCAuth `mapstructure:"auth"`
}

// gRPC auth modes. See GRPCAuth.
const (
	GRPCAuthToken        = "token"
	GRPCAuthSharedSecret = "shared_secret"
	GRPCAuthNone         = "none"
)

// GRPCAuth decides what a gRPC call must prove before a handler runs.
//
//	token         the same access token HTTP uses, checked against the same session store, so
//	              handlers get the same user context either way
//	shared_secret service-to-service calls with no user behind them
//	none          nothing is checked; only defensible when the port is reachable in-cluster
//	              only, which is the D5 argument applied to gRPC
type GRPCAuth struct {
	Mode   string `mapstructure:"mode"`
	Secret string `mapstructure:"secret"`
	// PublicMethods are exempt, as full method names ("/pkg.Service/Method") or a whole
	// service ("/pkg.Service/*"). Health checks and similar belong here.
	PublicMethods []string `mapstructure:"public_methods"`
}

// ModeOrDefault returns grpc.auth.mode, or GRPCAuthToken when unset. The default is the strict
// one: a port that answers before it has been told what to require should refuse, not serve.
func (a GRPCAuth) ModeOrDefault() string {
	mode := strings.ToLower(strings.TrimSpace(a.Mode))
	if mode == "" {
		return GRPCAuthToken
	}
	return mode
}

// GRPCTLS terminates TLS on the gRPC port itself, for deployments where it leaves the cluster.
type GRPCTLS struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
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
	// BodyLimit caps the request body in bytes. Zero means the default below, not "no limit".
	BodyLimit int `mapstructure:"body_limit"`
	// TrustedProxies lists the CIDRs or addresses allowed to set forwarding headers. Empty
	// means trust nobody, so the client IP is always the peer that actually connected.
	TrustedProxies []string `mapstructure:"trusted_proxies"`
	// ProxyHeader names the header carrying the real client IP, honoured only from a trusted
	// proxy. Defaults to X-Forwarded-For.
	ProxyHeader string `mapstructure:"proxy_header"`
	// HSTS adds Strict-Transport-Security. Only enable it where TLS terminates in front of
	// the app, otherwise browsers are told to refuse plain HTTP they still need.
	HSTS bool `mapstructure:"hsts"`
}

// Server timeout defaults. A server with no read timeout keeps a connection open for as long
// as a client is willing to dribble bytes at it, which is all a slowloris attack needs, so
// these apply whenever the value is left unset rather than defaulting to "no limit".
const (
	DefaultServerReadTimeout  = 15 * time.Second
	DefaultServerWriteTimeout = 15 * time.Second
	DefaultServerIdleTimeout  = 60 * time.Second

	// DefaultServerBodyLimit caps a request body at 10MB when server.body_limit is unset.
	//
	// It used to be hardcoded at 100MB, which is an invitation to exhaust memory on a service
	// whose own filesystem.max_file_size defaults to 200KB: Fiber buffers the body before a
	// handler sees it, so the limit is reached long before any upload validation runs. Raise
	// it deliberately if a route genuinely accepts large uploads.
	DefaultServerBodyLimit = 10 * 1024 * 1024
)

// BodyLimitOrDefault returns the configured body limit, or the 10MB default. A negative value
// is treated as unset rather than passed through, because Fiber reads a negative limit as "no
// limit at all".
func (s Server) BodyLimitOrDefault() int {
	if s.BodyLimit > 0 {
		return s.BodyLimit
	}
	return DefaultServerBodyLimit
}

// ProxyHeaderOrDefault returns the configured proxy header, or X-Forwarded-For.
func (s Server) ProxyHeaderOrDefault() string {
	if header := strings.TrimSpace(s.ProxyHeader); header != "" {
		return header
	}
	return "X-Forwarded-For"
}

// ReadTimeoutOrDefault returns server.read_timeout, or DefaultServerReadTimeout when unset.
func (s Server) ReadTimeoutOrDefault() time.Duration {
	if s.ReadTimeout > 0 {
		return s.ReadTimeout
	}
	return DefaultServerReadTimeout
}

// WriteTimeoutOrDefault returns server.write_timeout, or DefaultServerWriteTimeout when unset.
func (s Server) WriteTimeoutOrDefault() time.Duration {
	if s.WriteTimeout > 0 {
		return s.WriteTimeout
	}
	return DefaultServerWriteTimeout
}

// IdleTimeoutOrDefault returns server.idle_timeout, or DefaultServerIdleTimeout when unset.
func (s Server) IdleTimeoutOrDefault() time.Duration {
	if s.IdleTimeout > 0 {
		return s.IdleTimeout
	}
	return DefaultServerIdleTimeout
}

// CORS configures the browser cross-origin policy. It is read only when server.enable_cors is
// true; a service with no browser client should leave it off, because CORS relaxes a default
// that exists for a reason.
type CORS struct {
	// AllowOrigin is a comma-separated list of origins, or "*". Each entry must carry a scheme
	// (https://app.example.com), not a bare host.
	AllowOrigin  string `mapstructure:"allow_origin"`
	AllowMethods string `mapstructure:"allow_methods"`
	AllowHeaders string `mapstructure:"allow_headers"`
	// AllowCredentials lets the browser send cookies and Authorization on cross-origin
	// requests. It cannot be combined with AllowOrigin "*" — see validateServer.
	AllowCredentials bool `mapstructure:"allow_credentials"`
	// ExposeHeaders names response headers JavaScript may read. Without it a browser sees only
	// the CORS-safelisted ones, whatever the server sent.
	ExposeHeaders string `mapstructure:"expose_headers"`
	// MaxAge is how long a browser may cache the preflight result, in seconds.
	MaxAge int `mapstructure:"max_age"`
}

// AllowMethodsOrDefault returns the configured methods, or the set this API actually uses.
// Fiber's own default omits PATCH, which would make every PATCH route fail its preflight for
// no reason a reader of the config could see.
func (c CORS) AllowMethodsOrDefault() string {
	if strings.TrimSpace(c.AllowMethods) != "" {
		return c.AllowMethods
	}
	return "GET,POST,PUT,PATCH,DELETE,OPTIONS"
}

type DB struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	// MinOpenConnections is the pool's warm floor (database/sql: SetMaxIdleConns).
	MinOpenConnections    int `mapstructure:"min_open_connections"`
	MaxOpenConnections    int `mapstructure:"max_open_connections"`
	ConnectionMaxLifetime int `mapstructure:"connection_max_lifetime"`
	ConnectionMaxIdleTime int `mapstructure:"connection_max_idle_time"`
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
	DefaultLockoutMaxAttempts = 5
	DefaultLockoutDuration    = 10 * time.Minute
)

type Auth struct {
	SessionCache       string        `mapstructure:"session_cache"`        // auto | none | memory | redis (also used for the permission cache)
	SessionCacheTTL    time.Duration `mapstructure:"session_cache_ttl"`    // memory: max cross-instance revocation lag; redis: bounds staleness after direct DB edits
	PermissionCacheTTL time.Duration `mapstructure:"permission_cache_ttl"` // safety net; permission changes invalidate explicitly
	Revocation         string        `mapstructure:"revocation"`           // strict | refresh_only
	SessionExpiry      time.Duration `mapstructure:"session_expiry"`       // absolute session lifetime; also the refresh token lifetime
	RememberMeExpiry   time.Duration `mapstructure:"remember_me_expiry"`   // absolute session lifetime when remember_me = true
	RefreshReuseGrace  time.Duration `mapstructure:"refresh_reuse_grace"`  // window in which the just-replaced refresh token is still accepted
	Lockout            Lockout       `mapstructure:"lockout"`
	// RefreshTransport is where the refresh token travels: body (default; JSON body and
	// Authorization header, for mobile and server clients) or cookie (HttpOnly cookie, for
	// browsers, so page JavaScript can never read it).
	RefreshTransport string        `mapstructure:"refresh_transport"`
	RefreshCookie    RefreshCookie `mapstructure:"refresh_cookie"`
}

// Refresh token transports for auth.refresh_transport.
const (
	RefreshTransportBody   = "body"
	RefreshTransportCookie = "cookie"
)

// RefreshCookie configures the refresh cookie in cookie mode. Secure is not configurable: it is
// on everywhere except a local environment (see Config.IsLocal), because a refresh token sent
// over plain HTTP is a stolen refresh token.
type RefreshCookie struct {
	// SameSite is strict (default), lax, or none. strict and lax require the frontend and the
	// API to be same-site; a frontend on another site needs none, which also requires CORS with
	// credentials and explicit origins.
	SameSite string `mapstructure:"same_site"`
}

// UsesRefreshCookie reports whether auth.refresh_transport is cookie.
func (a Auth) UsesRefreshCookie() bool {
	return strings.EqualFold(strings.TrimSpace(a.RefreshTransport), RefreshTransportCookie)
}

// Lockout bounds password guessing: after MaxAttempts consecutive failures the account stops
// accepting any password, right or wrong, for Duration.
type Lockout struct {
	MaxAttempts int           `mapstructure:"max_attempts"`
	Duration    time.Duration `mapstructure:"duration"`
}

// MaxAttemptsOrDefault returns lockout.max_attempts, or DefaultLockoutMaxAttempts when unset.
func (l Lockout) MaxAttemptsOrDefault() int {
	if l.MaxAttempts > 0 {
		return l.MaxAttempts
	}
	return DefaultLockoutMaxAttempts
}

// DurationOrDefault returns lockout.duration, or DefaultLockoutDuration when unset.
func (l Lockout) DurationOrDefault() time.Duration {
	if l.Duration > 0 {
		return l.Duration
	}
	return DefaultLockoutDuration
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

// Crypto holds the key pkg/crypto encrypts with. Nothing in the boilerplate encrypts anything
// yet, so this key is a seam rather than a setting in use: read it and pass it to
// crypto.EncryptString when you add a column that needs it.
//
// It must be generated, not chosen. pkg/crypto derives the AES-256 key by taking SHA-256 of
// whatever string it is given, which is a single unsalted hash — it does not stretch a
// human-memorable passphrase into something resistant to an offline guess.
//
//	openssl rand -base64 32
type Crypto struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

// Service describes an upstream this application calls. It is a deliberate extension point:
// the map is keyed by a name of your choosing, so a deployment adds an upstream in config
// rather than in a new struct. Nothing in the boilerplate reads it — pkg/httpclient and
// pkg/grpcclient take a base URL from the caller — so an empty `service:` block is correct.
type Service struct {
	Name    string `mapstructure:"name"`
	BaseURL string `mapstructure:"base_url"`
	// Apikey is sent to that upstream, not accepted from it. Inbound partner keys are the
	// separate top-level `api_key` map.
	Apikey string `mapstructure:"api_key"`
}
