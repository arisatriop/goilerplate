package router

import "github.com/gofiber/fiber/v2"

func (r *RouteRegistry) upload(router fiber.Router) {
	upload := router.Group("/upload")

	// Multiple file upload
	upload.Post("/multiple", r.Wired.Handlers.Upload.UploadMultipleFiles)

	// Delete file
	upload.Delete("/", r.Wired.Handlers.Upload.DeleteFile)
}
