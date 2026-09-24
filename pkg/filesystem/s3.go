package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// immutableCacheControl is set on objects served through a CDN. Keys are random per upload, so
// an object at a given URL never changes and may be cached for good.
const immutableCacheControl = "public, max-age=31536000, immutable"

// S3Storage implements Storage for AWS S3 and S3-compatible stores. The bucket stays private:
// files are reached through a CDN in front of it, or through short-lived presigned URLs.
type S3Storage struct {
	client        *s3.Client
	presigner     *s3.PresignClient
	bucket        string
	cdnBaseURL    string
	presignExpiry time.Duration
}

// NewS3Storage creates an S3 storage. ctx is used only to load the AWS configuration.
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	loadOptions := []func(*config.LoadOptions) error{config.WithRegion(cfg.Region)}
	// Static keys only when both are given; otherwise the default chain, which is what the
	// example config always promised and the code never did.
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		loadOptions = append(loadOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(strings.TrimSuffix(cfg.Endpoint, "/"))
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &S3Storage{
		client:        client,
		presigner:     s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		cdnBaseURL:    strings.TrimSuffix(cfg.CDNBaseURL, "/"),
		presignExpiry: cfg.PresignExpiryOrDefault(),
	}, nil
}

// Upload uploads a multipart file. Its size is known from the form, so no request is spent
// asking S3 for it afterwards.
func (s *S3Storage) Upload(ctx context.Context, file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	if err := validateUpload(file.Size, file.Header.Get("Content-Type"), opts); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("opening upload: %w", err)
	}
	defer func() { _ = src.Close() }()

	filename := opts.Filename
	if filename == "" {
		filename = file.Filename
	}
	return s.put(ctx, src, file.Size, filename, opts)
}

// UploadFromReader uploads from any reader. The SDK needs a seekable body of known length to
// sign and retry the request (and refuses an unseekable one over plain HTTP, as MinIO often
// runs), so a reader that cannot seek is buffered first — bounded by server.body_limit when it
// comes from a request.
func (s *S3Storage) UploadFromReader(ctx context.Context, reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	seeker, ok := reader.(io.ReadSeeker)
	if !ok {
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("reading upload: %w", err)
		}
		seeker = bytes.NewReader(data)
	}

	size, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("measuring upload: %w", err)
	}
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewinding upload: %w", err)
	}
	return s.put(ctx, seeker, size, filename, opts)
}

func (s *S3Storage) put(ctx context.Context, body io.ReadSeeker, size int64, filename string, opts UploadOptions) (*UploadResult, error) {
	key := path.Join(strings.ReplaceAll(opts.Path, "\\", "/"), generateFilename(filename))
	key = strings.TrimPrefix(key, "/")
	mimeType := detectMimeType(filename)

	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(mimeType),
	}
	if s.cdnBaseURL != "" {
		input.CacheControl = aws.String(immutableCacheControl)
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return nil, fmt.Errorf("uploading to s3: %w", err)
	}

	fileURL, err := s.URL(ctx, key)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		OriginalName: filename,
		Filename:     filename,
		Path:         key,
		Size:         size,
		MimeType:     mimeType,
		URL:          fileURL,
		Driver:       DriverS3,
	}, nil
}

// Delete removes an object. Deleting a key that does not exist succeeds, as S3 itself does.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("deleting from s3: %w", err)
	}
	return nil
}

// Exists reports whether an object exists. A missing object is recognised by the SDK's typed
// error (or a 404 from a store that returns no body on HEAD), never by matching message text.
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	var notFound *types.NotFound
	var responseErr *awshttp.ResponseError
	if errors.As(err, &notFound) || (errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("checking s3 object: %w", err)
}

// URL returns where a client can fetch the object: the CDN address when one is configured,
// otherwise a presigned GET valid for presign_expiry.
func (s *S3Storage) URL(ctx context.Context, key string) (string, error) {
	if s.cdnBaseURL != "" {
		return s.cdnBaseURL + "/" + escapeKey(key), nil
	}

	presigned, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignExpiry))
	if err != nil {
		return "", fmt.Errorf("presigning s3 url: %w", err)
	}
	return presigned.URL, nil
}

// GetDriver returns the driver name
func (s *S3Storage) GetDriver() Driver {
	return DriverS3
}

// escapeKey percent-encodes each path segment of a key for use in a URL, keeping the slashes.
func escapeKey(key string) string {
	segments := strings.Split(key, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
