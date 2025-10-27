package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a simple UUID-like string

func GenerateUUID() string {
	return uuid.New().String()
}
