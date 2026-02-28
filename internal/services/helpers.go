package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// newID generates a random prefixed ID e.g. "run_a3f9c2...".
func newID(prefix string) string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("newID rand: %v", err))
	}
	return prefix + "_" + hex.EncodeToString(b)
}
