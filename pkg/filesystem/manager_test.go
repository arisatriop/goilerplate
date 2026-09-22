package filesystem

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyStorage records what the Manager forwards. Manager is a pass-through, so the only thing
// worth asserting is that it passes exactly what it was given to exactly the right method —
// a real backend would make that harder to see, not easier.
type spyStorage struct {
	uploadedName string
	uploadedOpts UploadOptions
	uploadedBody string
	deletedPath  string
	checkedPath  string
	urlPath      string

	err    error
	exists bool
}

func (s *spyStorage) Upload(file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	s.uploadedName = file.Filename
	s.uploadedOpts = opts
	return &UploadResult{Filename: file.Filename}, s.err
}

func (s *spyStorage) UploadFromReader(reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	body, _ := io.ReadAll(reader)
	s.uploadedBody = string(body)
	s.uploadedName = filename
	s.uploadedOpts = opts
	return &UploadResult{Filename: filename}, s.err
}

func (s *spyStorage) Delete(path string) error {
	s.deletedPath = path
	return s.err
}

func (s *spyStorage) Exists(path string) (bool, error) {
	s.checkedPath = path
	return s.exists, s.err
}

func (s *spyStorage) URL(path string) (string, error) {
	s.urlPath = path
	return "https://cdn.example.com/" + path, s.err
}

func (s *spyStorage) GetDriver() Driver { return DriverS3 }

// NewManager is the injection seam: it is what lets a caller substitute Storage in a test
// instead of reaching for a real bucket.
func TestManager_ForwardsToTheInjectedStorage(t *testing.T) {
	// Arrange
	spy := &spyStorage{exists: true}
	manager := NewManager(spy)
	opts := UploadOptions{Path: "avatars", MaxSize: 1024}

	// Act & Assert — upload
	result, err := manager.UploadFromReader(strings.NewReader("file body"), "photo.png", opts)
	require.NoError(t, err)
	assert.Equal(t, "photo.png", result.Filename)
	assert.Equal(t, "file body", spy.uploadedBody)
	assert.Equal(t, opts, spy.uploadedOpts, "options must arrive unmodified")

	// Act & Assert — delete
	require.NoError(t, manager.Delete("avatars/photo.png"))
	assert.Equal(t, "avatars/photo.png", spy.deletedPath)

	// Act & Assert — exists
	exists, err := manager.Exists("avatars/photo.png")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "avatars/photo.png", spy.checkedPath)

	// Act & Assert — url and driver
	url, err := manager.URL("avatars/photo.png")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/avatars/photo.png", url)
	assert.Equal(t, DriverS3, manager.GetDriver())
}

func TestManager_PropagatesTheStorageError(t *testing.T) {
	// Arrange
	spy := &spyStorage{err: errors.New("bucket unreachable")}
	manager := NewManager(spy)

	// Act & Assert
	_, err := manager.UploadFromReader(strings.NewReader("x"), "a.png", UploadOptions{})
	assert.Error(t, err)
	assert.Error(t, manager.Delete("a.png"))

	_, err = manager.Exists("a.png")
	assert.Error(t, err)

	_, err = manager.URL("a.png")
	assert.Error(t, err)
}

func TestNewManagerFromConfig_UnknownDriverIsAnError(t *testing.T) {
	_, err := NewManagerFromConfig(context.Background(), Config{Driver: "carrier-pigeon"})
	assert.Error(t, err)
}

func TestNewManagerFromConfig_Local(t *testing.T) {
	// Arrange
	root := t.TempDir()

	// Act
	manager, err := NewManagerFromConfig(context.Background(), Config{
		Driver: DriverLocal,
		Local:  LocalConfig{BasePath: root, BaseURL: "https://files.example.com"},
	})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, DriverLocal, manager.GetDriver())
}

// ── The local driver, end to end on a real temp directory ────────────────────

func TestLocalStorage_UploadThenExistsThenDelete(t *testing.T) {
	// Arrange
	root := t.TempDir()
	storage := NewLocalStorage(root, "https://files.example.com/")

	// Act — upload into a nested path that does not exist yet
	result, err := storage.UploadFromReader(
		bytes.NewBufferString("hello"), "note.txt", UploadOptions{Path: "docs/2026"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(5), result.Size)
	assert.Equal(t, "text/plain", result.MimeType)
	assert.Equal(t, DriverLocal, result.Driver)
	assert.Equal(t, filepath.Join("docs/2026", "note.txt"), result.Path)
	assert.Equal(t, "https://files.example.com/docs/2026/note.txt", result.URL,
		"the base URL's trailing slash must not double up")

	written, err := os.ReadFile(filepath.Join(root, "docs", "2026", "note.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(written))

	// Act & Assert — exists
	exists, err := storage.Exists("docs/2026/note.txt")
	require.NoError(t, err)
	assert.True(t, exists)

	missing, err := storage.Exists("docs/2026/absent.txt")
	require.NoError(t, err)
	assert.False(t, missing, "a missing file is not an error, it is a false")

	// Act & Assert — delete
	require.NoError(t, storage.Delete("docs/2026/note.txt"))

	exists, err = storage.Exists("docs/2026/note.txt")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestLocalStorage_DeletingSomethingThatIsNotThereIsAnError(t *testing.T) {
	storage := NewLocalStorage(t.TempDir(), "")

	assert.Error(t, storage.Delete("never-existed.txt"))
}

func TestLocalStorage_URLRequiresABaseURL(t *testing.T) {
	// Arrange
	withBase := NewLocalStorage(t.TempDir(), "https://files.example.com/")
	withoutBase := NewLocalStorage(t.TempDir(), "")

	// Act & Assert
	url, err := withBase.URL("/docs/note.txt")
	require.NoError(t, err)
	assert.Equal(t, "https://files.example.com/docs/note.txt", url)

	_, err = withoutBase.URL("docs/note.txt")
	assert.Error(t, err, "a URL with no base configured would be a broken link, not a blank one")
}

// An upload with no explicit path lands at the storage root rather than anywhere else.
func TestLocalStorage_EmptyPathUploadsToTheRoot(t *testing.T) {
	root := t.TempDir()
	storage := NewLocalStorage(root, "")

	result, err := storage.UploadFromReader(strings.NewReader("x"), "top.txt", UploadOptions{})

	require.NoError(t, err)
	assert.Equal(t, "top.txt", result.Path)
	assert.FileExists(t, filepath.Join(root, "top.txt"))
}
