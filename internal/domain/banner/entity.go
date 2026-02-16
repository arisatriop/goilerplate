package banner

import "github.com/google/uuid"

type Banner struct {
	ID          uuid.UUID
	StoreID     uuid.UUID
	Filetype    string
	FileStorage string
	Filename    string
	Filepath    string
	FileURL     *string
	IsActive    bool
}
