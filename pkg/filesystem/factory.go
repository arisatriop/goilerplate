package filesystem

import (
	"context"
	"fmt"
)

// StorageFactory is an interface for creating storage instances
type StorageFactory interface {
	Create(ctx context.Context, driver Driver, cfg Config) (Storage, error)
}

// DefaultStorageFactory implements StorageFactory
type DefaultStorageFactory struct{}

// NewStorageFactory creates a new storage factory
func NewStorageFactory() StorageFactory {
	return &DefaultStorageFactory{}
}

// Create creates a storage instance based on driver type
func (f *DefaultStorageFactory) Create(ctx context.Context, driver Driver, cfg Config) (Storage, error) {
	switch driver {
	case DriverLocal:
		return f.createLocal(cfg.Local)
	case DriverS3:
		return f.createS3(ctx, cfg.S3)
	case DriverDrive:
		return f.createDrive(ctx, cfg.Drive)
	default:
		return nil, &UnsupportedDriverError{Driver: string(driver)}
	}
}

func (f *DefaultStorageFactory) createLocal(cfg LocalConfig) (Storage, error) {
	if cfg.BasePath == "" {
		return nil, fmt.Errorf("local storage requires base_path")
	}
	return NewLocalStorage(cfg.BasePath, cfg.BaseURL), nil
}

func (f *DefaultStorageFactory) createS3(ctx context.Context, cfg S3Config) (Storage, error) {
	if cfg.Bucket == "" || cfg.Region == "" {
		return nil, fmt.Errorf("s3 storage requires bucket and region")
	}
	return NewS3Storage(ctx, cfg)
}

func (f *DefaultStorageFactory) createDrive(ctx context.Context, cfg DriveConfig) (Storage, error) {
	// Validate folder ID
	if cfg.FolderID == "" {
		return nil, fmt.Errorf("drive storage requires folder_id")
	}

	// Validate authentication: either Service Account OR OAuth
	hasServiceAccount := cfg.CredentialsFile != ""
	hasOAuth := cfg.ClientID != "" && cfg.ClientSecret != "" && cfg.RefreshToken != ""

	if !hasServiceAccount && !hasOAuth {
		return nil, fmt.Errorf("drive storage requires either credentials_file (service account) or client_id+client_secret+refresh_token (oauth)")
	}

	return NewDriveStorage(ctx, cfg)
}

// UnsupportedDriverError represents an unsupported driver error
type UnsupportedDriverError struct {
	Driver string
}

func (e *UnsupportedDriverError) Error() string {
	return fmt.Sprintf("unsupported storage driver: %s", e.Driver)
}
