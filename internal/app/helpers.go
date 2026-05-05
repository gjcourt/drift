package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// newID generates a random prefixed ID e.g. "run_a3f9c2...". It returns an
// error rather than panicking when crypto/rand fails (FIPS / sandbox /
// /dev/urandom unavailable), so callers can surface it like any other I/O
// error in this codebase.
func newID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("newID rand: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(b), nil
}
