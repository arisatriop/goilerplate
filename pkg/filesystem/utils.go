package filesystem

import (
	"fmt"
	"goilerplate/pkg/utils"
	"net/http"
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
		return utils.Error(http.StatusBadRequest, fmt.Sprintf("Ukuran file melebihi batas maksimum %.0fmb", convertToMB(opts.MaxSize)-1))
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
			return utils.Error(http.StatusBadRequest, fmt.Sprintf("mime type %s is not allowed", mimeType))
		}
	}

	return nil
}

// convertToMB converts bytes to megabytes
func convertToMB(bytes int64) float64 {
	return float64(bytes) / (1024 * 1024)
}
