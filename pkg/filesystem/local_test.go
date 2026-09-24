package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Upload paths and file names arrive from request data, so a path that climbs out of the
// storage root must be refused rather than cleaned and followed.
func TestResolveWithin_RejectsEscapes(t *testing.T) {
	root := t.TempDir()

	escapes := map[string][]string{
		"parent":             {"..", "passwd"},
		"nested parent":      {"uploads", "..", "..", "passwd"},
		"parent in filename": {"uploads", "../../passwd"},
		"root itself":        {".."},
	}

	for name, parts := range escapes {
		t.Run(name, func(t *testing.T) {
			_, err := resolveWithin(root, parts...)
			assert.Error(t, err, "%v should not resolve", parts)
		})
	}
}

// A sibling directory sharing the root's prefix is outside the root. A string-prefix check
// would wave /storage-evil through for a root of /storage.
func TestResolveWithin_RejectsSiblingSharingThePrefix(t *testing.T) {
	root := filepath.Join(t.TempDir(), "storage")
	require.NoError(t, os.MkdirAll(root, 0750))

	_, err := resolveWithin(root, "../storage-evil/file.txt")

	assert.Error(t, err)
}

// An absolute part is contained rather than rejected: filepath.Join treats everything after
// the first element as relative, so "/etc/passwd" lands at <root>/etc/passwd. Recorded as a
// test because it is the non-obvious half of why this helper is safe.
func TestResolveWithin_ContainsAbsoluteParts(t *testing.T) {
	root := t.TempDir()

	full, err := resolveWithin(root, "/etc/passwd")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "etc", "passwd"), full)
}

func TestResolveWithin_AllowsPathsInsideTheRoot(t *testing.T) {
	root := t.TempDir()

	full, err := resolveWithin(root, "avatars", "2026", "photo.png")

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(full, root), "%q should sit under %q", full, root)
	assert.Equal(t, filepath.Join(root, "avatars", "2026", "photo.png"), full)
}

func TestLocalStorage_UploadRefusesToWriteOutsideTheRoot(t *testing.T) {
	root := t.TempDir()
	storage := NewLocalStorage(root, "")

	_, err := storage.UploadFromReader(t.Context(), strings.NewReader("payload"), "../escaped.txt", UploadOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "escapes the storage root")
	_, statErr := os.Stat(filepath.Join(filepath.Dir(root), "escaped.txt"))
	assert.True(t, os.IsNotExist(statErr), "nothing should have been written outside the root")
}

func TestLocalStorage_DeleteRefusesPathsOutsideTheRoot(t *testing.T) {
	root := t.TempDir()
	victim := filepath.Join(filepath.Dir(root), "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("keep me"), 0600))
	t.Cleanup(func() { _ = os.Remove(victim) })

	err := NewLocalStorage(root, "").Delete(t.Context(), "../victim.txt")

	require.Error(t, err)
	assert.FileExists(t, victim, "the file outside the root must still be there")
}
