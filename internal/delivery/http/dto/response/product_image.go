package dtoresponse

type ProductImageUploadResponse struct {
	Filetype    string `json:"fileType"`
	FileStorage string `json:"fileStorage"`
	Filename    string `json:"fileName"`
	Filepath    string `json:"filePath"`
	FileURL     string `json:"fileUrl"`
}
