package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Config struct {
	Secret        string
	Issuer        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type UserClaims struct {
	SessionID  string
	UserID     string
	Email      string
	Role       string
	IsVerified bool
}

type Claims struct {
	UserID     string `json:"user_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Role       string `json:"role,omitempty"`
	IsVerified bool   `json:"is_verified,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	TokenType  string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type Tokenizer interface {
	GenerateTokenPair(payload UserClaims) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	RefreshExpiry() time.Duration
}

type tokenizer struct {
	secret        []byte
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewTokenizer creates and validates a new Tokenizer instance.
func NewTokenizer(cfg Config) (Tokenizer, error) {
	if len(cfg.Secret) < 32 {
		return nil, errors.New("jwt secret must be at least 32 bytes")
	}

	accessExpiry := cfg.AccessExpiry
	if accessExpiry <= 0 {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry := cfg.RefreshExpiry
	if refreshExpiry <= 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}

	return &tokenizer{
		secret:        []byte(cfg.Secret),
		issuer:        cfg.Issuer,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

func (t *tokenizer) RefreshExpiry() time.Duration {
	return t.refreshExpiry
}

func (t *tokenizer) GenerateTokenPair(payload UserClaims) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		UserID:     payload.UserID,
		Email:      payload.Email,
		Role:       payload.Role,
		IsVerified: payload.IsVerified,
		SessionID:  payload.SessionID,
		TokenType:  TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   payload.UserID,
			Issuer:    t.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessExpiry)),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(t.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := Claims{
		UserID:    payload.UserID,
		SessionID: payload.SessionID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   payload.UserID,
			Issuer:    t.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.refreshExpiry)),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(t.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(t.accessExpiry.Seconds()),
	}, nil
}

func (t *tokenizer) validate(tokenString, expectedType string) (*Claims, error) {
	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	}
	if t.issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(t.issuer))
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return t.secret, nil
	}, parserOptions...)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("%w: %v", ErrExpiredToken, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%w: expected token_type %q, got %q", ErrInvalidToken, expectedType, claims.TokenType)
	}

	return claims, nil
}

func (t *tokenizer) ValidateAccessToken(tokenString string) (*Claims, error) {
	return t.validate(tokenString, TokenTypeAccess)
}

func (t *tokenizer) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return t.validate(tokenString, TokenTypeRefresh)
}

// HashToken hashes a token using SHA-256 for secure storage in Redis.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
