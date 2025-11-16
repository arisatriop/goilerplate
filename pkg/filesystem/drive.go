package filesystem

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveStorage implements Storage for Google Drive
type DriveStorage struct {
	service      *drive.Service
	folderID     string
	ctx          context.Context
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

		// Log token info
		// isValid, expiresIn := tokenManager.TokenInfo()
		// if isValid {
		// 	fmt.Printf("OAuth token loaded successfully. Expires in: %v\n", expiresIn.Round(time.Second))
		// }
	} else if cfg.CredentialsFile != "" {
		// Use Service Account authentication (no token management needed)
		useOAuth = false
		service, err = drive.NewService(ctx, option.WithCredentialsFile(cfg.CredentialsFile))
		if err != nil {
			return nil, fmt.Errorf("failed to create drive service with service account: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no authentication credentials provided: either provide OAuth credentials (client_id, client_secret, refresh_token) or service account credentials_file")
	}

	return &DriveStorage{
		service:      service,
		folderID:     cfg.FolderID,
		ctx:          ctx,
		tokenManager: tokenManager,
		useOAuth:     useOAuth,
	}, nil
}

// Upload uploads file from multipart form
func (d *DriveStorage) Upload(file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	if err := validateUpload(file.Size, file.Header.Get("Content-Type"), opts); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Preserve original filename
	originalName := file.Filename

	// Generate new filename if not provided
	filename := opts.Filename
	if filename == "" {
		filename = generateFilename(originalName)
	}

	// Pass file metadata to avoid extra API calls
	result, err := d.uploadWithMetadata(src, filename, file.Size, file.Header.Get("Content-Type"), opts)
	if err != nil {
		return nil, err
	}

	// Set the correct original name
	result.OriginalName = originalName

	return result, nil
}

// UploadFromReader uploads from io.Reader (without known size/mimeType)
func (d *DriveStorage) UploadFromReader(reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	return d.uploadWithMetadata(reader, filename, 0, "", opts)
}

// uploadWithMetadata performs the actual upload with known metadata to avoid extra API calls
func (d *DriveStorage) uploadWithMetadata(reader io.Reader, filename string, fileSize int64, mimeType string, opts UploadOptions) (*UploadResult, error) {
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
		Context(d.ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload to drive: %w", err)
	}

	// Make public asynchronously (don't block upload response)
	if opts.Public {
		go func(fileID string) {
			permission := &drive.Permission{
				Type: "anyone",
				Role: "reader",
			}
			_, err := d.service.Permissions.Create(fileID, permission).SupportsAllDrives(true).Context(d.ctx).Do()
			if err != nil {
				// Log warning but don't fail the upload
				fmt.Printf("Warning: failed to set public permission for %s: %v\n", fileID, err)
			}
		}(driveFile.Id)
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
func (d *DriveStorage) Delete(fileID string) error {
	if err := d.service.Files.Delete(fileID).SupportsAllDrives(true).Context(d.ctx).Do(); err != nil {
		return fmt.Errorf("failed to delete from drive: %w", err)
	}
	return nil
}

// Exists checks if file exists
func (d *DriveStorage) Exists(fileID string) (bool, error) {
	_, err := d.service.Files.Get(fileID).SupportsAllDrives(true).Context(d.ctx).Do()
	if err != nil {
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	return true, nil
}

// URL gets direct view URL for file
func (d *DriveStorage) URL(fileID string) (string, error) {
	return fmt.Sprintf("https://drive.google.com/uc?export=view&id=%s", fileID), nil
}

// GetDriver returns the driver name
func (d *DriveStorage) GetDriver() Driver {
	return DriverDrive
}
