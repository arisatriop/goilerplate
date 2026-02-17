package userrole

import "github.com/google/uuid"

type UserRole struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}
