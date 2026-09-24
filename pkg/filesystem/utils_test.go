package filesystem

import (
	"goilerplate/pkg/apperr"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 bytes"},
		{512, "512 bytes"},
		{1024, "1KB"},
		// The default filesystem.max_file_size. The previous formatter reported this as
		// "-1mb", because it converted to megabytes and then subtracted one.
		{200 * 1024, "200KB"},
		{1536, "1.5KB"},
		{1 << 20, "1MB"},
		{10 * (1 << 20), "10MB"},
		{100 * (1 << 20), "100MB"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, humanSize(tt.bytes))
	}
}

func TestValidateUpload_Size(t *testing.T) {
	// Arrange
	opts := UploadOptions{MaxSize: 200 * 1024}

	// Act
	err := validateUpload(200*1024+1, "image/png", opts)

	// Assert
	appErr, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.Invalid, appErr.Kind)
	assert.Equal(t, "file_too_large", appErr.Code)
	assert.Equal(t, "File exceeds the maximum size of 200KB", appErr.Error())

	assert.NoError(t, validateUpload(200*1024, "image/png", opts), "exactly the limit is allowed")
}

// MaxSize 0 means no limit, so a caller that does not configure one is not silently capped.
func TestValidateUpload_ZeroMaxSizeMeansNoLimit(t *testing.T) {
	assert.NoError(t, validateUpload(1<<30, "image/png", UploadOptions{}))
}

func TestValidateUpload_MimeType(t *testing.T) {
	opts := UploadOptions{AllowedMimeTypes: []string{"image/png", "image/jpeg"}}

	assert.NoError(t, validateUpload(10, "image/png", opts))
	assert.NoError(t, validateUpload(10, "image/jpeg", opts))

	err := validateUpload(10, "application/x-msdownload", opts)
	appErr, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, "file_type_not_allowed", appErr.Code)

	// An empty allowlist means every type is accepted, which is what an unconfigured caller
	// gets — worth pinning so it is a decision rather than a surprise.
	assert.NoError(t, validateUpload(10, "application/x-msdownload", UploadOptions{}))
}

func TestDetectMimeType(t *testing.T) {
	tests := map[string]string{
		"photo.png":      "image/png",
		"photo.PNG":      "image/png",
		"photo.jpg":      "image/jpeg",
		"photo.jpeg":     "image/jpeg",
		"report.pdf":     "application/pdf",
		"data.csv":       "text/csv",
		"archive.zip":    "application/zip",
		"notes.txt":      "text/plain",
		"no-extension":   "application/octet-stream",
		"unknown.xyz":    "application/octet-stream",
		"archive.tar.gz": "application/octet-stream",
		"UPPER.TXT":      "text/plain",
		"weird.name.doc": "application/msword",
	}

	for filename, want := range tests {
		assert.Equal(t, want, detectMimeType(filename), filename)
	}
}

// The generated name must not collide and must keep the extension, because the extension is
// what detectMimeType and every downstream consumer reads.
func TestGenerateFilename(t *testing.T) {
	first := generateFilename("holiday photo.png")
	second := generateFilename("holiday photo.png")

	assert.NotEqual(t, first, second, "two uploads of one name must not overwrite each other")
	assert.True(t, len(first) > 4 && first[len(first)-4:] == ".png", "extension preserved: %s", first)
	assert.NotContains(t, first, " ", "the original name's spaces must not survive into a path")
}
