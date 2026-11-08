package otp

import (
	"crypto/hmac"
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

// Hash hashes an OTP code using HMAC-SHA256 bound to the recipient email.
func Hash(code, secret, email string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(email))
	mac.Write([]byte(":"))
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks if the provided OTP code matches the hashed OTP using constant-time comparison.
func Verify(code, hash, secret, email string) bool {
	codeHash := Hash(code, secret, email)
	return subtle.ConstantTimeCompare([]byte(codeHash), []byte(hash)) == 1
}
