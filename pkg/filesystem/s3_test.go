package filesystem

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBucket = "test-bucket"

// fakeS3 is just enough of the S3 API to watch what the driver sends. It records every request
// and answers HEAD by key: "present/..." exists, "denied/..." is 403, anything else is 404.
type fakeS3 struct {
	mu       sync.Mutex
	requests []*http.Request
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	f.mu.Lock()
	f.requests = append(f.requests, r)
	f.mu.Unlock()

	switch r.Method {
	case http.MethodPut:
		w.Header().Set("ETag", `"etag"`)
		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodHead:
		switch {
		case strings.Contains(r.URL.Path, "/present/"):
			w.Header().Set("Content-Length", "3")
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/denied/"):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (f *fakeS3) only(t *testing.T) *http.Request {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	require.Len(t, f.requests, 1)
	return f.requests[0]
}

// newFakeS3Storage points the driver at a fake S3-compatible store, addressed path-style as
// MinIO is.
func newFakeS3Storage(t *testing.T, mutate func(*S3Config)) (*S3Storage, *fakeS3, string) {
	t.Helper()

	fake := &fakeS3{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	cfg := S3Config{
		AccessKeyID:     "AKIATESTKEY",
		SecretAccessKey: "test-secret",
		Region:          "us-east-1",
		Bucket:          testBucket,
		Endpoint:        server.URL,
		UsePathStyle:    true,
	}
	if mutate != nil {
		mutate(&cfg)
	}

	storage, err := NewS3Storage(t.Context(), cfg)
	require.NoError(t, err)
	return storage, fake, server.URL
}

// The old driver sent Bucket "" on every PutObject and no bucket at all on Delete/Exists, so it
// could not upload to AWS; and it set a public-read ACL, which new buckets reject.
func TestS3_UploadWritesIntoTheBucketWithoutAnACL(t *testing.T) {
	storage, fake, endpoint := newFakeS3Storage(t, nil)

	result, err := storage.UploadFromReader(t.Context(), strings.NewReader("png-bytes"), "photo.png",
		UploadOptions{Path: "avatars", Public: true})

	require.NoError(t, err)
	req := fake.only(t)
	assert.Equal(t, http.MethodPut, req.Method)
	assert.True(t, strings.HasPrefix(req.URL.Path, "/"+testBucket+"/avatars/"), "path-style: /bucket/key, got %s", req.URL.Path)
	assert.Empty(t, req.Header.Get("X-Amz-Acl"), "objects stay private; access goes through a CDN or a presigned URL")
	assert.Equal(t, "image/png", req.Header.Get("Content-Type"))
	assert.Empty(t, req.Header.Get("Cache-Control"), "no CDN, so no long-lived caching")

	assert.Equal(t, int64(len("png-bytes")), result.Size, "the size comes from the upload, not from a HeadObject")
	assert.True(t, strings.HasPrefix(result.Path, "avatars/"))
	assert.Equal(t, DriverS3, result.Driver)

	presigned, err := url.Parse(result.URL)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimPrefix(endpoint, "http://"), presigned.Host)
	assert.Equal(t, "3600", presigned.Query().Get("X-Amz-Expires"), "a private object is handed out as a 1h presigned URL")
	assert.NotEmpty(t, presigned.Query().Get("X-Amz-Signature"))
}

// A reader that cannot seek is buffered so the SDK gets a body of known length; over plain HTTP
// it would otherwise refuse to send it.
func TestS3_UploadFromANonSeekableReader(t *testing.T) {
	storage, fake, _ := newFakeS3Storage(t, nil)

	result, err := storage.UploadFromReader(t.Context(), io.MultiReader(strings.NewReader("abc"), strings.NewReader("def")),
		"notes.txt", UploadOptions{})

	require.NoError(t, err)
	assert.Equal(t, int64(6), result.Size)
	assert.Equal(t, http.MethodPut, fake.only(t).Method)
}

func TestS3_BehindACDNObjectsAreImmutableAndServedFromIt(t *testing.T) {
	storage, fake, _ := newFakeS3Storage(t, func(c *S3Config) { c.CDNBaseURL = "https://cdn.example.com/" })

	result, err := storage.UploadFromReader(t.Context(), strings.NewReader("x"), "a b.webp", UploadOptions{Path: "img"})

	require.NoError(t, err)
	assert.Equal(t, immutableCacheControl, fake.only(t).Header.Get("Cache-Control"))
	assert.Equal(t, "image/webp", fake.only(t).Header.Get("Content-Type"))
	assert.True(t, strings.HasPrefix(result.URL, "https://cdn.example.com/img/"), result.URL)
	assert.NotContains(t, result.URL, "X-Amz-Signature", "a CDN URL is permanent, not presigned")
}

func TestS3_PresignExpiryIsConfigurable(t *testing.T) {
	storage, _, _ := newFakeS3Storage(t, func(c *S3Config) { c.PresignExpiry = 10 * time.Minute })

	link, err := storage.URL(t.Context(), "docs/a.pdf")

	require.NoError(t, err)
	parsed, err := url.Parse(link)
	require.NoError(t, err)
	assert.Equal(t, "600", parsed.Query().Get("X-Amz-Expires"))
}

// Exists used to decide "not found" by searching the error message for "NotFound" or "404".
func TestS3_Exists(t *testing.T) {
	storage, _, _ := newFakeS3Storage(t, nil)

	present, err := storage.Exists(t.Context(), "present/a.txt")
	require.NoError(t, err)
	assert.True(t, present)

	missing, err := storage.Exists(t.Context(), "absent/a.txt")
	require.NoError(t, err)
	assert.False(t, missing, "a missing object is an answer, not an error")

	_, err = storage.Exists(t.Context(), "denied/a.txt")
	assert.Error(t, err, "a 403 is a real failure and must not read as 'does not exist'")
}

func TestS3_DeleteAddressesTheBucket(t *testing.T) {
	storage, fake, _ := newFakeS3Storage(t, nil)

	require.NoError(t, storage.Delete(t.Context(), "docs/a.pdf"))

	req := fake.only(t)
	assert.Equal(t, http.MethodDelete, req.Method)
	assert.Equal(t, "/"+testBucket+"/docs/a.pdf", req.URL.Path)
}

// Against AWS itself the bucket goes in the host by default, or in the path when asked. Presigning
// is computed locally, so this checks the addressing without a network.
func TestS3_AWSAddressing(t *testing.T) {
	tests := []struct {
		name      string
		pathStyle bool
		wantHost  string
		wantPath  string
	}{
		{"virtual-hosted", false, testBucket + ".s3.eu-west-1.amazonaws.com", "/docs/a.pdf"},
		{"path style", true, "s3.eu-west-1.amazonaws.com", "/" + testBucket + "/docs/a.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewS3Storage(t.Context(), S3Config{
				AccessKeyID: "AKIATESTKEY", SecretAccessKey: "s", Region: "eu-west-1", Bucket: testBucket,
				UsePathStyle: tt.pathStyle,
			})
			require.NoError(t, err)

			link, err := storage.URL(t.Context(), "docs/a.pdf")
			require.NoError(t, err)
			parsed, err := url.Parse(link)
			require.NoError(t, err)

			assert.Equal(t, tt.wantHost, parsed.Host)
			assert.Equal(t, tt.wantPath, parsed.Path)
		})
	}
}

// Empty keys mean the default credential chain — what the example config always promised. The
// old driver built a static provider from the two empty strings instead.
func TestS3_EmptyKeysUseTheDefaultCredentialChain(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAFROMENVCHAIN")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "env-secret")
	t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/none")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/none")

	storage, err := NewS3Storage(t.Context(), S3Config{Region: "us-east-1", Bucket: testBucket})
	require.NoError(t, err)

	link, err := storage.URL(t.Context(), "a.txt")
	require.NoError(t, err)
	parsed, err := url.Parse(link)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(parsed.Query().Get("X-Amz-Credential"), "AKIAFROMENVCHAIN/"))
}

func TestEscapeKey_KeepsSlashesAndEncodesSegments(t *testing.T) {
	assert.Equal(t, "img/a%20b%3F.png", escapeKey("img/a b?.png"))
}
