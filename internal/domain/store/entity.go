package store

import (
	"fmt"
	"goilerplate/pkg/crypto"

	"github.com/google/uuid"
)

type Store struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Name     string
	Desc     *string
	Address  *string
	Phone    *string
	Email    *string
	WebURL   string
	IsActive bool
}

func (s *Store) GenerateWebURL(key string, baseURL string) {
	encryptedStoreID, err := crypto.EncryptString(s.ID.String(), key)
	if err != nil {
		// Fallback to raw UUID if encryption fails
		// In production, you might want to log this error
		encryptedStoreID = s.ID.String()
	}

	s.WebURL = fmt.Sprintf("%s/%s", baseURL, encryptedStoreID)
}
