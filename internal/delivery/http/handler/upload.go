package handler

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/filesystem"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Upload struct {
	validator     *validator.Validate
	filesystemMgr *filesystem.Manager
	maxFileSize   int64
}

func NewUpload(validator *validator.Validate, filesystemMgr *filesystem.Manager, maxFileSize int64) *Upload {
	return &Upload{
		validator:     validator,
		filesystemMgr: filesystemMgr,
		maxFileSize:   maxFileSize,
	}
}

// UploadFile handles single file upload
func (h *Upload) UploadFile(ctx *fiber.Ctx) error {
	// Parse multipart form
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.Locals("error_detail", err.Error())
		return response.BadRequest(ctx, "File is required", nil)
	}

	// Parse upload options from form
	// Build upload options - server controls everything
	opts := filesystem.UploadOptions{
		Path:     "", // Empty = root folder
		Filename: "", // Empty = auto-generate
		Public:   true,
	}

	// Upload file
	result, err := h.filesystemMgr.Upload(file, opts)
	if err != nil {
		// Log the error for debugging
		ctx.Locals("upload_error", err.Error())
		ctx.Locals("file_name", file.Filename)
		ctx.Locals("file_size", file.Size)
		return response.HandleError(ctx, err)
	}

	// Map to response DTO
	responseData := &dtoresponse.UploadFileResponse{
		OriginalName: result.OriginalName,
		Filename:     result.Filename,
		Path:         result.Path,
		Size:         result.Size,
		MimeType:     result.MimeType,
		URL:          result.URL,
		Driver:       string(result.Driver),
	}

	return response.Success(ctx, responseData, response.WithMessage("File uploaded successfully"))
}

// UploadMultipleFiles handles multiple file uploads
func (h *Upload) UploadMultipleFiles(ctx *fiber.Ctx) error {
	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return response.BadRequest(ctx, "At least one file is required", nil)
	}

	// Parse upload options from form
	// Build upload options - server controls everything
	opts := filesystem.UploadOptions{
		Path:     "trash", // Empty = root folder
		Filename: "",      // Empty = auto-generate
		Public:   true,
		MaxSize:  h.maxFileSize,
	}

	// Upload all files
	var results []*dtoresponse.UploadFileResponse
	for _, file := range files {
		result, err := h.filesystemMgr.Upload(file, opts)
		if err != nil {
			return response.HandleError(ctx, err)
		}

		results = append(results, &dtoresponse.UploadFileResponse{
			OriginalName: result.OriginalName,
			Filename:     result.Filename,
			Path:         result.Path,
			Size:         result.Size,
			MimeType:     result.MimeType,
			URL:          result.URL,
			Driver:       string(result.Driver),
		})
	}

	if len(results) == 0 {
		return response.BadRequest(ctx, "Failed to upload any file", nil)
	}

	return response.Success(ctx, results, response.WithMessage("Files uploaded successfully"))
}

// DeleteFile handles file deletion
func (h *Upload) DeleteFile(ctx *fiber.Ctx) error {
	path := ctx.Query("path")
	if path == "" {
		return response.BadRequest(ctx, "Path parameter is required", nil)
	}

	if err := h.filesystemMgr.Delete(path); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil, response.WithMessage("File deleted successfully"))
}

// GetFileURL gets public URL for a file
func (h *Upload) GetFileURL(ctx *fiber.Ctx) error {
	path := ctx.Query("path")
	if path == "" {
		return response.BadRequest(ctx, "Path parameter is required", nil)
	}

	url, err := h.filesystemMgr.URL(path)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, fiber.Map{
		"path": path,
		"url":  url,
	}, response.WithMessage("URL retrieved successfully"))
}

// CheckFileExists checks if a file exists
func (h *Upload) CheckFileExists(ctx *fiber.Ctx) error {
	path := ctx.Query("path")
	if path == "" {
		return response.BadRequest(ctx, "Path parameter is required", nil)
	}

	exists, err := h.filesystemMgr.Exists(path)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, fiber.Map{
		"path":   path,
		"exists": exists,
	}, response.WithMessage("File existence checked successfully"))
}
