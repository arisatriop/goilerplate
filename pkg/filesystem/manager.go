package filesystem

import (
	"context"
	"io"
	"mime/multipart"
)

// Manager manages file storage operations with dependency injection
type Manager struct {
	storage Storage
	factory StorageFactory
}

// NewManager creates a manager with injected dependencies
func NewManager(storage Storage) *Manager {
	return &Manager{
		storage: storage,
		factory: NewStorageFactory(),
	}
}

// NewManagerFromConfig creates a manager from configuration
func NewManagerFromConfig(ctx context.Context, cfg Config) (*Manager, error) {
	factory := NewStorageFactory()
	storage, err := factory.Create(ctx, cfg.Driver, cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{
		storage: storage,
		factory: factory,
	}, nil
}

// Upload uploads a file
func (m *Manager) Upload(file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	return m.storage.Upload(file, opts)
}

// UploadFromReader uploads from reader
func (m *Manager) UploadFromReader(reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	return m.storage.UploadFromReader(reader, filename, opts)
}

// Delete deletes a file
func (m *Manager) Delete(path string) error {
	return m.storage.Delete(path)
}

// Exists checks if file exists
func (m *Manager) Exists(path string) (bool, error) {
	return m.storage.Exists(path)
}

// URL gets public URL
func (m *Manager) URL(path string) (string, error) {
	return m.storage.URL(path)
}

// GetDriver returns current driver
func (m *Manager) GetDriver() Driver {
	return m.storage.GetDriver()
}
