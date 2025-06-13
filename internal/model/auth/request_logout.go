package auth

import "github.com/google/uuid"

type LogoutRequest struct {
	ID uuid.UUID
}
