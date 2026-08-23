package internal

import (
	"crypto/rand"
	"encoding/hex"
)

func NewID() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}
