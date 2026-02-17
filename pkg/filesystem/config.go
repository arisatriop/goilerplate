package filesystem

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
type S3Config struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	Endpoint        string `mapstructure:"endpoint"`
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
