package filesystem

import (
	"fmt"
	"goilerplate/pkg/apperr"
	"goilerplate/pkg/utils"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// generateFilename creates a unique filename with format: YYYYMMDD_HHMMSS_randomhash.ext
func generateFilename(original string) string {
	ext := filepath.Ext(original)
	now := utils.Now()
	dateTime := now.Format("20060102_150405")
	randomHash := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s%s", dateTime, randomHash, ext)
}

// detectMimeType detects MIME type from filename extension
func detectMimeType(filename string) string {
	types := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".zip":  "application/zip",
	}
	if mime, ok := types[strings.ToLower(filepath.Ext(filename))]; ok {
		return mime
	}
	return "application/octet-stream"
}

// validateUpload validates file upload options
func validateUpload(fileSize int64, mimeType string, opts UploadOptions) error {
	// Validate size
	if opts.MaxSize > 0 && fileSize > opts.MaxSize {
		return apperr.New(apperr.Invalid, "file_too_large",
			fmt.Sprintf("File exceeds the maximum size of %s", humanSize(opts.MaxSize)))
	}

	// Validate MIME type
	if len(opts.AllowedMimeTypes) > 0 {
		allowed := false
		for _, mt := range opts.AllowedMimeTypes {
			if mt == mimeType {
				allowed = true
				break
			}
		}
		if !allowed {
			return apperr.New(apperr.Invalid, "file_type_not_allowed", fmt.Sprintf("File type %s is not allowed", mimeType))
		}
	}

	return nil
}

// humanSize renders a byte count the way the limit was written in config.
//
// It replaces fmt.Sprintf("%.0fmb", convertToMB(max)-1), which was wrong twice over. The -1
// was presumably meant to round down, but it subtracts a whole megabyte: the default limit of
// 200KB is 0.195MB, so the message read "melebihi batas maksimum -1mb". And it reported every
// limit in megabytes, so any limit under 1MB rendered as 0 or a negative number.
func humanSize(bytes int64) string {
	const (
		kb = 1 << 10
		mb = 1 << 20
	)

	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.6gMB", float64(bytes)/mb)
	case bytes >= kb:
		return fmt.Sprintf("%.6gKB", float64(bytes)/kb)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
