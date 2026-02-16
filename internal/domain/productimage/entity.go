package productimage

import "github.com/google/uuid"

type ProductImage struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	FileType    string
	FileStorage string
	FileName    string
	FilePath    string
	FileURL     string
	IsPrimary   bool
	IsActive    bool
}
