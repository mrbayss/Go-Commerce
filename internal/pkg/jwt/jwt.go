package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
	SessionID  string `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func GenerateAccessToken(sessionID, userID, email, role string, isVerified bool, secret string, expiry time.Duration, issuer string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		Email:      email,
		Role:       role,
		IsVerified: isVerified,
		SessionID:  sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(sessionID, userID, secret string, expiry time.Duration, issuer string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ID:        sessionID,
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateTokenPair(sessionID, userID, email, role string, isVerified bool, secret string, accessExpiry, refreshExpiry time.Duration, issuer string) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(sessionID, userID, email, role, isVerified, secret, accessExpiry, issuer)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(sessionID, userID, secret, refreshExpiry, issuer)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessExpiry.Seconds()),
	}, nil
}

// HashToken hashes a token using SHA-256 for secure storage in Redis.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ValidateRefreshToken(tokenString, secret string) (userID string, sessionID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", "", ErrExpiredToken
		}
		return "", "", ErrInvalidToken
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", "", ErrInvalidToken
	}
	return claims.Subject, claims.ID, nil
}

func RefreshAccessToken(refreshToken, secret string, accessExpiry time.Duration, issuer string) (string, error) {
	userID, sessionID, err := ValidateRefreshToken(refreshToken, secret)
	if err != nil {
		return "", err
	}
	// Need to fetch user details from DB to generate new access token with full claims
	// This is a simplified version - in practice, fetch user from DB
	return GenerateAccessToken(sessionID, userID, "", "", false, secret, accessExpiry, issuer)
}