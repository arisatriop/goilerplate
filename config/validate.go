package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// MinSecretBytes is the minimum length of a JWT signing secret (256 bits for HS256).
	MinSecretBytes = 32

	// minSecretDistinctBytes rejects low-entropy secrets such as "aaaa..." in production.
	minSecretDistinctBytes = 10

	envProduction = "production"
)

// placeholderPattern matches unreplaced template values such as <JWT_ACCESS_SECRET_KEY>.
var placeholderPattern = regexp.MustCompile(`^<[^<>]*>$`)

// sampleSecretMarkers are fragments of example values that must never reach production.
var sampleSecretMarkers = []string{"changeme", "change-me", "change_me", "your_", "your-", "example", "placeholder", "supersecret"}

// Validate checks the configuration for missing or unsafe values and returns every
// problem found, joined into one error. It performs no I/O and must run before any
// connection is opened.
func (c *Config) Validate() error {
	v := &validation{}

	c.validateApp(v)
	c.validateServer(v)
	c.validateDB(v)
	c.validateRedis(v)
	c.validateGRPC(v)
	c.validateOTel(v)
	c.validateJWT(v)
	c.validateAuth(v)
	c.validateFileSystem(v)
	c.validateAPIKeys(v)

	return v.err()
}

// IsProduction reports whether the app runs with app.env=production.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.App.Env), envProduction)
}

func (c *Config) validateApp(v *validation) {
	v.required("app.env", c.App.Env)
	v.required("app.name", c.App.Name)
}

func (c *Config) validateServer(v *validation) {
	v.port("server.port", c.Server.Port)
}

func (c *Config) validateDB(v *validation) {
	v.required("db.host", c.DB.Host)
	v.port("db.port", c.DB.Port)
	v.required("db.name", c.DB.Name)
	v.required("db.username", c.DB.Username)
	v.notPlaceholder("db.password", c.DB.Password)
	if c.DB.MaxOpenConnections < 1 {
		v.addf("db.max_open_connections must be at least 1, got %d", c.DB.MaxOpenConnections)
	}
}

func (c *Config) validateRedis(v *validation) {
	if !c.Redis.Enabled {
		return
	}
	v.required("redis.host", c.Redis.Host)
	v.notPlaceholder("redis.password", c.Redis.Password)
}

func (c *Config) validateGRPC(v *validation) {
	if !c.GRPC.Enabled {
		return
	}
	v.port("grpc.port", c.GRPC.Port)
	if c.GRPC.Port == c.Server.Port {
		v.addf("grpc.port must differ from server.port (both are %d)", c.GRPC.Port)
	}
}

func (c *Config) validateOTel(v *validation) {
	if !c.OTel.Enabled {
		return
	}
	v.required("otel.endpoint", c.OTel.Endpoint)
}

func (c *Config) validateJWT(v *validation) {
	jwt := c.JWT
	production := c.IsProduction()

	v.required("jwt.key_id", jwt.KeyID)
	v.secret("jwt.access_secret", jwt.AccessSecret, MinSecretBytes, production)
	v.secret("jwt.refresh_secret", jwt.RefreshSecret, MinSecretBytes, production)
	if jwt.AccessSecret != "" && jwt.AccessSecret == jwt.RefreshSecret {
		v.addf("jwt.access_secret and jwt.refresh_secret must be different")
	}

	v.required("jwt.issuer", jwt.Issuer)
	v.required("jwt.audience", jwt.Audience)
	if jwt.AccessTokenExpiry <= 0 {
		v.addf("jwt.access_token_expiry must be greater than 0")
	}
	if jwt.Leeway < 0 {
		v.addf("jwt.leeway must not be negative")
	}

	// A duplicate key id would make "kid" ambiguous, so a token could be verified with the
	// wrong secret depending on map order.
	seen := map[string]bool{jwt.KeyID: true}
	for i, key := range jwt.PreviousKeys {
		field := fmt.Sprintf("jwt.previous_keys[%d]", i)

		v.required(field+".key_id", key.KeyID)
		if key.KeyID != "" && seen[key.KeyID] {
			v.addf("%s.key_id %q is already used by another key", field, key.KeyID)
		}
		seen[key.KeyID] = true

		v.secret(field+".access_secret", key.AccessSecret, MinSecretBytes, production)
		v.secret(field+".refresh_secret", key.RefreshSecret, MinSecretBytes, production)
	}
}

func (c *Config) validateAuth(v *validation) {
	switch strings.ToLower(strings.TrimSpace(c.Auth.SessionCache)) {
	case "", CacheModeAuto, CacheModeNone, CacheModeMemory:
	case CacheModeRedis:
		if !c.Redis.Enabled {
			v.addf("auth.session_cache=redis requires redis.enabled=true")
		}
	default:
		v.addf("auth.session_cache must be auto, none, memory, or redis, got %q", c.Auth.SessionCache)
	}

	if c.Auth.SessionCacheTTL < 0 {
		v.addf("auth.session_cache_ttl must not be negative")
	}
	if c.Auth.PermissionCacheTTL < 0 {
		v.addf("auth.permission_cache_ttl must not be negative")
	}

	// A session must outlive the access tokens issued within it, otherwise a client holding a
	// valid access token would already have an unusable session.
	sessionExpiry := c.Auth.SessionExpiryOrDefault()
	if c.Auth.SessionExpiry < 0 {
		v.addf("auth.session_expiry must not be negative")
	} else if sessionExpiry <= c.JWT.AccessTokenExpiry {
		v.addf("auth.session_expiry must be greater than jwt.access_token_expiry")
	}

	if c.Auth.RememberMeExpiry < 0 {
		v.addf("auth.remember_me_expiry must not be negative")
	} else if c.Auth.RememberMeExpiryOrDefault() < sessionExpiry {
		v.addf("auth.remember_me_expiry must not be shorter than auth.session_expiry")
	}
}

func (c *Config) validateFileSystem(v *validation) {
	fs := c.FileSystem

	switch strings.ToLower(fs.Driver) {
	case "local":
		v.required("filesystem.local.base_path", fs.Local.BasePath)
	case "s3":
		v.required("filesystem.s3.bucket", fs.S3.Bucket)
		v.required("filesystem.s3.region", fs.S3.Region)
		v.notPlaceholder("filesystem.s3.access_key_id", fs.S3.AccessKeyID)
		v.notPlaceholder("filesystem.s3.secret_access_key", fs.S3.SecretAccessKey)
		if c.IsProduction() && fs.S3.SecretAccessKey != "" {
			v.notSample("filesystem.s3.secret_access_key", fs.S3.SecretAccessKey)
		}
	case "drive":
		hasOAuth := fs.Drive.ClientID != "" && fs.Drive.ClientSecret != "" && fs.Drive.RefreshToken != ""
		if fs.Drive.CredentialsFile == "" && !hasOAuth {
			v.addf("filesystem.drive requires credentials_file, or client_id, client_secret, and refresh_token")
		}
	default:
		v.addf("filesystem.driver must be local, s3, or drive, got %q", fs.Driver)
	}

	if fs.MaxFileSize <= 0 {
		v.addf("filesystem.max_file_size must be greater than 0")
	}
}

func (c *Config) validateAPIKeys(v *validation) {
	for name, key := range c.Apikeys {
		field := "api_key." + name
		v.required(field, key)
		if c.IsProduction() {
			v.notSample(field, key)
		}
	}
}

// validation collects configuration problems so they can be reported together.
type validation struct {
	problems []string
}

func (v *validation) addf(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

// required rejects empty values and unreplaced placeholders.
func (v *validation) required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.addf("%s is required", field)
		return
	}
	v.notPlaceholder(field, value)
}

// notPlaceholder rejects unreplaced placeholders but allows empty values.
func (v *validation) notPlaceholder(field, value string) {
	if placeholderPattern.MatchString(strings.TrimSpace(value)) {
		v.addf("%s still contains the placeholder %s", field, value)
	}
}

func (v *validation) port(field string, port int) {
	if port < 1 || port > 65535 {
		v.addf("%s must be between 1 and 65535, got %d", field, port)
	}
}

// secret validates a signing secret. Secret values are never included in messages.
func (v *validation) secret(field, value string, minBytes int, production bool) {
	if value == "" {
		v.addf("%s is required", field)
		return
	}
	if placeholderPattern.MatchString(strings.TrimSpace(value)) {
		v.addf("%s still contains a placeholder", field)
		return
	}
	if len(value) < minBytes {
		v.addf("%s must be at least %d bytes, got %d", field, minBytes, len(value))
	}
	if production {
		v.notSample(field, value)
	}
}

// notSample rejects example or low-entropy values. Secret values are never included in messages.
func (v *validation) notSample(field, value string) {
	lower := strings.ToLower(value)
	for _, marker := range sampleSecretMarkers {
		if strings.Contains(lower, marker) {
			v.addf("%s looks like an example value (contains %q); use a randomly generated secret in production", field, marker)
			return
		}
	}

	distinct := make(map[byte]struct{})
	for i := 0; i < len(value); i++ {
		distinct[value[i]] = struct{}{}
	}
	if len(distinct) < minSecretDistinctBytes {
		v.addf("%s has too little variety (%d distinct characters); use a randomly generated secret in production", field, len(distinct))
	}
}

func (v *validation) err() error {
	if len(v.problems) == 0 {
		return nil
	}

	errs := make([]error, len(v.problems))
	for i, problem := range v.problems {
		errs[i] = errors.New(problem)
	}
	return errors.Join(errs...)
}
