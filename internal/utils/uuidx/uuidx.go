package uuidx

import (
	"strings"

	"github.com/google/uuid"
)

const (
	defaultShortLen = 8
	maxShortLen     = 32
)

// New returns a UUID string.
func New() string {
	return uuid.NewString()
}

// NewShort returns a short id based on UUID, by removing hyphens and truncating to n.
// It clamps n to [1, 32]; when n is invalid, it falls back to 8.
func NewShort(n int) string {
	if n <= 0 || n > maxShortLen {
		n = defaultShortLen
	}
	s := strings.ReplaceAll(uuid.NewString(), "-", "")
	return s[:n]
}

