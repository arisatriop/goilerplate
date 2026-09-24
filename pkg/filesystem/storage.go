package filesystem

import (
	"context"
	"io"
	"mime/multipart"
)

// Storage is the main interface for file storage operations
type Storage interface {
	Uploader
	Deleter
	Checker
	URLProvider
}

// Uploader handles file upload operations
type Uploader interface {
	Upload(ctx context.Context, file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error)
	UploadFromReader(ctx context.Context, reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error)
}

// Deleter handles file deletion
type Deleter interface {
	Delete(ctx context.Context, path string) error
}

// Checker checks file existence
type Checker interface {
	Exists(ctx context.Context, path string) (bool, error)
}

// URLProvider provides file URLs
type URLProvider interface {
	// URL returns where a client can fetch the file. Depending on the driver it is permanent
	// (local base URL, CDN) or time-limited (an S3 presigned URL), so it is handed out per request
	// rather than stored.
	URL(ctx context.Context, path string) (string, error)
	GetDriver() Driver
}

// Driver types
type Driver string

const (
	DriverLocal Driver = "local"
	DriverS3    Driver = "s3"
	DriverDrive Driver = "drive"
)

// UploadOptions contains file upload configuration
type UploadOptions struct {
	Path             string
	Filename         string
	MaxSize          int64
	AllowedMimeTypes []string
	// Public asks for the file to be readable without credentials. local: served from base_url.
	// drive: shared with "anyone with the link". s3: ignored — objects stay private, and are
	// reached through cdn_base_url or a presigned URL, because public ACLs are disabled on new
	// buckets and make every object's exposure a per-upload decision.
	Public bool
}

// UploadResult contains upload result information
type UploadResult struct {
	OriginalName string `json:"originalName"`
	Filename     string `json:"filename"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mimeType"`
	URL          string `json:"url"` // Preview URL for embedding (iframe-friendly)
	Driver       Driver `json:"driver"`
}
