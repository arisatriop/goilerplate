package filesystem

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements Storage for local filesystem
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Upload uploads file from multipart form
func (l *LocalStorage) Upload(file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	// Validate
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

	result, err := l.UploadFromReader(src, filename, opts)
	if err != nil {
		return nil, err
	}

	// Set the correct original name
	result.OriginalName = originalName

	return result, nil
}

// UploadFromReader uploads from io.Reader
func (l *LocalStorage) UploadFromReader(reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	fullPath := filepath.Join(l.basePath, opts.Path)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	destPath := filepath.Join(fullPath, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	relativePath := filepath.Join(opts.Path, filename)
	url := ""
	if l.baseURL != "" {
		url = strings.TrimSuffix(l.baseURL, "/") + "/" + strings.TrimPrefix(relativePath, "/")
	}

	return &UploadResult{
		OriginalName: filename,
		Filename:     filename,
		Path:         relativePath,
		Size:         size,
		MimeType:     detectMimeType(filename),
		URL:          url,
		Driver:       DriverLocal,
	}, nil
}

// Delete deletes a file
func (l *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(l.basePath, path)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Exists checks if file exists
func (l *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(l.basePath, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	return true, nil
}

// URL gets public URL for file
func (l *LocalStorage) URL(path string) (string, error) {
	if l.baseURL == "" {
		return "", fmt.Errorf("base URL not configured")
	}
	return strings.TrimSuffix(l.baseURL, "/") + "/" + strings.TrimPrefix(path, "/"), nil
}

// GetDriver returns the driver name
func (l *LocalStorage) GetDriver() Driver {
	return DriverLocal
}
