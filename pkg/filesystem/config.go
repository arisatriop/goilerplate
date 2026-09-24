package filesystem

import "time"

// Config holds filesystem configuration
type Config struct {
	Driver Driver
	Local  LocalConfig
	S3     S3Config
	Drive  DriveConfig
}

// LocalConfig for local filesystem storage
type LocalConfig struct {
	BasePath string `mapstructure:"base_path"`
	BaseURL  string `mapstructure:"base_url"`
}

// S3Config for AWS S3 storage
// S3Config configures the S3 driver, for AWS S3 and S3-compatible stores (MinIO, Cloudflare R2).
// Every field beyond bucket and region is optional and safe at its zero value.
type S3Config struct {
	// AccessKeyID and SecretAccessKey are static credentials. Leave both empty to use the
	// default AWS credential chain (environment, shared config, IAM role, IRSA) — the right
	// choice on AWS, where a long-lived key is one more secret to leak.
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	// Endpoint is the service endpoint of an S3-compatible store, without the bucket
	// (http://minio:9000, https://<account>.r2.cloudflarestorage.com). Empty means AWS.
	Endpoint string `mapstructure:"endpoint"`
	// UsePathStyle addresses objects as endpoint/bucket/key instead of bucket.endpoint/key.
	// MinIO needs it; AWS and R2 do not.
	UsePathStyle bool `mapstructure:"use_path_style"`
	// CDNBaseURL serves files through a CDN (CloudFront, Cloudflare) in front of a private
	// bucket. Objects are then written with a year-long immutable Cache-Control — safe because
	// every upload gets a new, random key. Empty means files are handed out as presigned URLs.
	CDNBaseURL string `mapstructure:"cdn_base_url"`
	// PresignExpiry is how long a presigned URL stays valid when there is no CDN. Default 1h;
	// SigV4 caps it at 7 days.
	PresignExpiry time.Duration `mapstructure:"presign_expiry"`
}

// DefaultPresignExpiry is used when s3.presign_expiry is unset.
const DefaultPresignExpiry = time.Hour

// MaxPresignExpiry is the longest validity SigV4 allows a presigned URL.
const MaxPresignExpiry = 7 * 24 * time.Hour

// PresignExpiryOrDefault returns presign_expiry, or DefaultPresignExpiry when unset.
func (c S3Config) PresignExpiryOrDefault() time.Duration {
	if c.PresignExpiry > 0 {
		return c.PresignExpiry
	}
	return DefaultPresignExpiry
}

// DriveConfig for Google Drive storage
type DriveConfig struct {
	// Service Account authentication
	CredentialsFile string `mapstructure:"credentials_file"`

	// OAuth2 authentication
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RefreshToken string `mapstructure:"refresh_token"`

	// Common config
	FolderID string `mapstructure:"folder_id"`
}
