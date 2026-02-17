package filesystem

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements Storage for AWS S3
type S3Storage struct {
	client   *s3.Client
	bucket   string
	region   string
	endpoint string
	ctx      context.Context
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID, cfg.SecretAccessKey, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			// For custom endpoints that include bucket name, don't use path style
			o.UsePathStyle = false
		}
	})

	return &S3Storage{
		client:   client,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		endpoint: cfg.Endpoint,
		ctx:      ctx,
	}, nil
}

// Upload uploads file from multipart form
func (s *S3Storage) Upload(file *multipart.FileHeader, opts UploadOptions) (*UploadResult, error) {
	if err := validateUpload(file.Size, file.Header.Get("Content-Type"), opts); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	filename := opts.Filename
	if filename == "" {
		filename = file.Filename
	}

	return s.UploadFromReader(src, filename, opts)
}

// UploadFromReader uploads from io.Reader
func (s *S3Storage) UploadFromReader(reader io.Reader, filename string, opts UploadOptions) (*UploadResult, error) {
	key := filepath.Join(opts.Path, generateFilename(filename))
	key = strings.ReplaceAll(key, "\\", "/")

	bucket := ""
	input := &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(detectMimeType(filename)),
	}

	if opts.Public {
		input.ACL = types.ObjectCannedACLPublicRead
	}

	_, err := s.client.PutObject(s.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Get file size
	head, _ := s.client.HeadObject(s.ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    aws.String(key),
	})

	size := int64(0)
	if head != nil && head.ContentLength != nil {
		size = *head.ContentLength
	}

	url := ""
	if opts.Public {
		if s.endpoint != "" {
			// Custom endpoint (MinIO, etc) - already includes bucket in domain
			url = fmt.Sprintf("%s/%s", s.endpoint, key)
		} else {
			// AWS S3 standard endpoint
			url = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
		}
	}

	return &UploadResult{
		OriginalName: filename,
		Filename:     filename,
		Path:         key,
		Size:         size,
		MimeType:     detectMimeType(filename),
		URL:          url,
		Driver:       DriverS3,
	}, nil
}

// Delete deletes a file
func (s *S3Storage) Delete(key string) error {
	// For custom endpoints that already include bucket name, use empty bucket

	_, err := s.client.DeleteObject(s.ctx, &s3.DeleteObjectInput{
		Key: aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

// Exists checks if file exists
func (s *S3Storage) Exists(key string) (bool, error) {
	// For custom endpoints that already include bucket name, use empty bucket
	_, err := s.client.HeadObject(s.ctx, &s3.HeadObjectInput{
		Key: aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file: %w", err)
	}
	return true, nil
}

// URL gets public URL for file
func (s *S3Storage) URL(key string) (string, error) {
	if s.endpoint != "" {
		// Custom endpoint (MinIO, etc) - already includes bucket in domain
		return fmt.Sprintf("%s/%s", s.endpoint, key), nil
	}
	// AWS S3 standard endpoint
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

// GetDriver returns the driver name
func (s *S3Storage) GetDriver() Driver {
	return DriverS3
}
