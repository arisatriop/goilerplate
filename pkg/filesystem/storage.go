package filesystem

import (
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
	Upload(file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error)
	UploadFromReader(reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error)
}

// Deleter handles file deletion
type Deleter interface {
	Delete(path string) error
}

// Checker checks file existence
type Checker interface {
	Exists(path string) (bool, error)
}

// URLProvider provides file URLs
type URLProvider interface {
	URL(path string) (string, error)
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
	Public           bool
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
