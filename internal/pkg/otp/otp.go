package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Generate creates a secure random 6-digit OTP code as a string.
func Generate() (string, error) {
	otp, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", otp.Int64()), nil
}

// Hash hashes an OTP code using SHA-256.
func Hash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// Verify checks if the provided OTP code matches the hashed OTP using constant-time comparison.
func Verify(code, hash string) bool {
	codeHash := Hash(code)
	return subtle.ConstantTimeCompare([]byte(codeHash), []byte(hash)) == 1
}
