package dtorequest

type CreateProductsRequest struct {
	Products []ProductRequest `json:"products" validate:"required,dive"`
}

type ProductRequest struct {
	Name        string         `json:"name" validate:"required"`
	Desc        *string        `json:"desc"` // pointer → allows null
	Price       string         `json:"price" validate:"required"`
	CategoryIDs []string       `json:"categoryIds" validate:"required,min=1,dive"`
	Images      []ProductImage `json:"images" validate:"required,min=1,dive"`
}

type ProductImagesRequest struct {
	Images []ProductImage `json:"images" validate:"required,min=1,dive"`
}

type ProductImage struct {
	FileType    string `json:"fileType" validate:"required"`
	FileStorage string `json:"fileStorage" validate:"required"`
	FileName    string `json:"fileName" validate:"required"`
	FilePath    string `json:"filePath" validate:"required"`
	FileURL     string `json:"fileUrl" validate:"required,url"`
	IsPrimary   bool   `json:"isPrimary"`
}

type UpdateProductRequest struct {
	Name  string  `json:"name" validate:"required"`
	Desc  *string `json:"desc"`
	Price string  `json:"price" validate:"required"`
}

type CreateProductCategoriesRequest struct {
	CategoryIDs []string `json:"categoryIds" validate:"required,min=1,dive"`
}
