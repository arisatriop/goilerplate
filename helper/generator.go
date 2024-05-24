package helper

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateShortUUID() string {
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	timestamp := time.Now().Unix()
	randomPart := randSource.Int63n(1e6) // Generates a random number with a maximum of 6 digits
	return fmt.Sprintf("%d%d", timestamp, randomPart)
}
