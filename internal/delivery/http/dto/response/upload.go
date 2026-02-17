package dtoresponse

type UploadFileResponse struct {
	OriginalName string `json:"originalName"`
	Filename     string `json:"fileName"`
	Path         string `json:"filePath"`
	Size         int64  `json:"fileSize"`
	URL          string `json:"fileUrl"` // Preview URL (iframe-friendly, works for PDF & images)
	MimeType     string `json:"mimeType"`
	Driver       string `json:"driver"`
}
