// Package slug generates URL-friendly, conflict-free slugs.
package slug

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrConflict signals that no free slug could be found after all retries.
var ErrConflict = errors.New("slug already taken")

const (
	maxAttempts = 5
	fallback    = "item"
)

// Generate converts a name to a lowercase ASCII slug:
// non-alphanumeric runs become "-", leading/trailing "-" are trimmed.
// An empty result falls back to "item".
func Generate(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case unicode.IsLetter(r) && r <= unicode.MaxASCII, unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	s := strings.Trim(b.String(), "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" {
		return fallback
	}
	return s
}

// Unique returns Generate(name) if free, otherwise appends a random
// 4-char hex suffix and retries up to maxAttempts times. It returns an
// error if no free slug is found (callers map it to 409).
func Unique(ctx context.Context, name string, exists func(context.Context, string) (bool, error)) (string, error) {
	base := Generate(name)
	slug := base
	for i := 0; i < maxAttempts; i++ {
		used, err := exists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !used {
			return slug, nil
		}
		slug = base + "-" + randomHex()
	}
	return "", fmt.Errorf("%w: %s", ErrConflict, base)
}

func randomHex() string {
	buf := make([]byte, 2)
	if _, err := rand.Read(buf); err != nil {
		return "0000"
	}
	return hex.EncodeToString(buf)
}
