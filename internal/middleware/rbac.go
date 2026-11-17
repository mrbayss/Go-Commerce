package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/entity"
)

type requireRolesMiddleware struct {
	allowed []entity.UserRole
}

func RequireRoles(allowed ...entity.UserRole) fiber.Handler {
	m := &requireRolesMiddleware{allowed: allowed}
	return m.handle
}

func (m *requireRolesMiddleware) handle(c fiber.Ctx) error {
	claims, ok := AuthClaimsFrom(c.Context())
	if !ok {
		return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}

	if len(m.allowed) == 0 {
		return &apperrors.AppError{Code: http.StatusForbidden, Message: "forbidden"}
	}

	for _, role := range m.allowed {
		if normalizeRole(claims.Role) == normalizeRole(string(role)) {
			return c.Next()
		}
	}
	return &apperrors.AppError{Code: http.StatusForbidden, Message: "forbidden"}
}

func normalizeRole(role string) string {
	return strings.ToLower(role)
}
