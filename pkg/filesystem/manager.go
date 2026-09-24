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
func (m *Manager) Upload(ctx context.Context, file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	return m.storage.Upload(ctx, file, opts)
}

// UploadFromReader uploads from reader
func (m *Manager) UploadFromReader(ctx context.Context, reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	return m.storage.UploadFromReader(ctx, reader, filename, opts)
}

// Delete deletes a file
func (m *Manager) Delete(ctx context.Context, path string) error {
	return m.storage.Delete(ctx, path)
}

// Exists checks if file exists
func (m *Manager) Exists(ctx context.Context, path string) (bool, error) {
	return m.storage.Exists(ctx, path)
}

// URL gets public URL
func (m *Manager) URL(ctx context.Context, path string) (string, error) {
	return m.storage.URL(ctx, path)
}

// GetDriver returns current driver
func (m *Manager) GetDriver() Driver {
	return m.storage.GetDriver()
}
