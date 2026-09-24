package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// DriveStorage implements Storage for Google Drive
type DriveStorage struct {
	service      *drive.Service
	folderID     string
	tokenManager *TokenManager // For OAuth token management
	useOAuth     bool          // Flag to indicate OAuth vs Service Account
}

// NewDriveStorage creates a new Drive storage instance
// Supports both OAuth2 and Service Account authentication
// If OAuth credentials (ClientID, ClientSecret, RefreshToken) are provided, uses OAuth2 with automatic token refresh
// Otherwise, falls back to Service Account authentication with CredentialsFile
func NewDriveStorage(ctx context.Context, cfg DriveConfig) (*DriveStorage, error) {
	var service *drive.Service
	var tokenManager *TokenManager
	var useOAuth bool
	var err error

	// Check if OAuth2 credentials are provided
	if cfg.ClientID != "" && cfg.ClientSecret != "" && cfg.RefreshToken != "" {
		// Use OAuth2 authentication with TokenManager for automatic refresh
		useOAuth = true

		// Create token manager with caching (cache in ./storage/cache directory)
		tokenManager, err = NewTokenManager(
			cfg.ClientID,
			cfg.ClientSecret,
			cfg.RefreshToken,
			"./storage/cache",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create token manager: %w", err)
		}

		// Get initial token (will use cached token if valid)
		token, err := tokenManager.GetToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get initial token: %w", err)
		}

		// Create OAuth config
		config := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{drive.DriveFileScope},
		}

		// Create HTTP client with token source that auto-refreshes
		tokenSource := config.TokenSource(ctx, token)
		client := oauth2.NewClient(ctx, tokenSource)

		service, err = drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("failed to create drive service with OAuth: %w", err)
		}
	} else if cfg.CredentialsFile != "" {
		// Use Service Account authentication (no token management needed).
		//
		// WithAuthCredentialsFile pins the credential type, where the deprecated
		// WithCredentialsFile accepted any of them. That matters here: the path comes from
		// config, so a file swapped for, say, an external-account configuration would
		// otherwise be loaded and honoured — and an external-account config can name an
		// arbitrary URL as its token source.
		useOAuth = false
		service, err = drive.NewService(ctx,
			option.WithAuthCredentialsFile(option.ServiceAccount, cfg.CredentialsFile))
		if err != nil {
			return nil, fmt.Errorf("failed to create drive service with service account: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no authentication credentials provided: either provide OAuth credentials (client_id, client_secret, refresh_token) or service account credentials_file")
	}

	return &DriveStorage{
		service:      service,
		folderID:     cfg.FolderID,
		tokenManager: tokenManager,
		useOAuth:     useOAuth,
	}, nil
}

// Upload uploads file from multipart form
func (d *DriveStorage) Upload(ctx context.Context, file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	if err := validateUpload(file.Size, file.Header.Get("Content-Type"), opts); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = src.Close() }()

	// Preserve original filename
	originalName := file.Filename

	// Generate new filename if not provided
	filename := opts.Filename
	if filename == "" {
		filename = generateFilename(originalName)
	}

	// Pass file metadata to avoid extra API calls
	result, err := d.uploadWithMetadata(ctx, src, filename, file.Size, file.Header.Get("Content-Type"), opts)
	if err != nil {
		return nil, err
	}

	// Set the correct original name
	result.OriginalName = originalName

	return result, nil
}

// UploadFromReader uploads from io.Reader (without known size/mimeType)
func (d *DriveStorage) UploadFromReader(ctx context.Context, reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	return d.uploadWithMetadata(ctx, reader, filename, 0, "", opts)
}

// uploadWithMetadata performs the actual upload with known metadata to avoid extra API calls
func (d *DriveStorage) uploadWithMetadata(ctx context.Context, reader io.Reader, filename string, fileSize int64, mimeType string, opts UploadOptions) (*UploadResult, error) {
	// Detect MIME type if not provided
	if mimeType == "" {
		mimeType = detectMimeType(filename)
	}

	fileMetadata := &drive.File{
		Name:    filename,
		Parents: []string{d.folderID},
	}

	driveFile, err := d.service.Files.Create(fileMetadata).
		Media(reader).
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload to drive: %w", err)
	}

	// Sharing happens before the upload is reported as done. It used to run in a detached
	// goroutine on a context stored at construction, so a failure left a private file behind a
	// URL the caller had been told was public, and nothing could wait for it at shutdown.
	if opts.Public {
		permission := &drive.Permission{Type: "anyone", Role: "reader"}
		if _, err := d.service.Permissions.Create(driveFile.Id, permission).SupportsAllDrives(true).Context(ctx).Do(); err != nil {
			// Remove the upload rather than leave an orphan the caller has no ID for.
			if delErr := d.Delete(ctx, driveFile.Id); delErr != nil {
				err = errors.Join(err, fmt.Errorf("removing the unshared upload: %w", delErr))
			}
			return nil, fmt.Errorf("sharing drive file: %w", err)
		}
	}

	// Use provided file size or 0 if unknown
	size := fileSize
	if size == 0 && driveFile.Size > 0 {
		size = driveFile.Size
	}

	// Return preview URL for iframe embedding (works for PDF, images, etc.)
	previewURL := fmt.Sprintf("https://drive.google.com/file/d/%s/preview", driveFile.Id)

	return &UploadResult{
		OriginalName: filename,
		Filename:     filename,
		Path:         driveFile.Id,
		Size:         size,
		MimeType:     mimeType,
		URL:          previewURL,
		Driver:       DriverDrive,
	}, nil
}

// Delete deletes a file
func (d *DriveStorage) Delete(ctx context.Context, fileID string) error {
	if err := d.service.Files.Delete(fileID).SupportsAllDrives(true).Context(ctx).Do(); err != nil {
		return fmt.Errorf("failed to delete from drive: %w", err)
	}
	return nil
}

// Exists checks if file exists
func (d *DriveStorage) Exists(ctx context.Context, fileID string) (bool, error) {
	_, err := d.service.Files.Get(fileID).SupportsAllDrives(true).Context(ctx).Do()
	if err == nil {
		return true, nil
	}
	// A missing file is an answer, not a failure. It used to be reported as an error, so
	// Exists could never return false.
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("checking drive file: %w", err)
}

// URL gets direct view URL for file
func (d *DriveStorage) URL(_ context.Context, fileID string) (string, error) {
	return fmt.Sprintf("https://drive.google.com/uc?export=view&id=%s", fileID), nil
}

// GetDriver returns the driver name
func (d *DriveStorage) GetDriver() Driver {
	return DriverDrive
}
