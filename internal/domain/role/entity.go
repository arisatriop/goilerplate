package role

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
}
