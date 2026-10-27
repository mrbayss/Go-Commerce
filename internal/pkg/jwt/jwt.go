package jwt

import (
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
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func GenerateAccessToken(userID, email, role string, isVerified bool, secret string, expiry time.Duration, issuer string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		Email:      email,
		Role:       role,
		IsVerified: isVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID, secret string, expiry time.Duration, issuer string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateTokenPair(userID, email, role string, isVerified bool, secret string, accessExpiry, refreshExpiry time.Duration, issuer string) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, email, role, isVerified, secret, accessExpiry, issuer)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(userID, secret, refreshExpiry, issuer)
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

func ValidateRefreshToken(tokenString, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}

func RefreshAccessToken(refreshToken, secret string, accessExpiry time.Duration, issuer string) (string, error) {
	userID, err := ValidateRefreshToken(refreshToken, secret)
	if err != nil {
		return "", err
	}
	// Need to fetch user details from DB to generate new access token with full claims
	// This is a simplified version - in practice, fetch user from DB
	return GenerateAccessToken(userID, "", "", false, secret, accessExpiry, issuer)
}