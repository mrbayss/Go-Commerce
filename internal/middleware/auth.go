// Package middleware: guard transport-level autentikasi (RequireAuth)
// dan RBAC (RequireRoles) — di-wire dari composition root.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/pkg/jwt"
	"github.com/redis/go-redis/v9"
)

type authClaimsKey struct{}

func WithAuthClaims(ctx context.Context, claims *jwt.Claims) context.Context {
	return context.WithValue(ctx, authClaimsKey{}, claims)
}

func AuthClaimsFrom(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(authClaimsKey{}).(*jwt.Claims)
	if !ok || claims == nil {
		return nil, false
	}
	return claims, true
}

type requireAuthMiddleware struct {
	tokenizer jwt.Tokenizer
	redis     *redis.Client
}

func RequireAuth(tokenizer jwt.Tokenizer, redis *redis.Client) fiber.Handler {
	m := &requireAuthMiddleware{
		tokenizer: tokenizer,
		redis:     redis,
	}
	return m.handle
}

func (m *requireAuthMiddleware) handle(c fiber.Ctx) error {
	token := extractBearerToken(c.Get("Authorization"))
	if token == "" {
		return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "missing access token"}
	}

	claims, err := m.tokenizer.ValidateAccessToken(token)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "access token expired"}
		}
		return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "invalid access token"}
	}

	if m.redis != nil && claims.SessionID != "" {
		blacklistKey := "blacklist:session:" + claims.SessionID
		exists, err := m.redis.Exists(c.Context(), blacklistKey).Result()
		if err != nil {
			return apperrors.Wrap(err, http.StatusBadGateway, "failed to verify token revocation")
		}
		if exists > 0 {
			return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "token has been revoked"}
		}
	}

	ctx := WithAuthClaims(c.Context(), claims)
	c.SetContext(ctx)
	return c.Next()
}

func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	const prefix = "bearer "
	if len(authHeader) < len(prefix) || !strings.EqualFold(authHeader[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(authHeader[len(prefix):])
}
