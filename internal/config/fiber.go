package config

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// NewFiber creates and configures a new Fiber application instance.
func NewFiber(config *viper.Viper, validate *FiberValidator, log *logrus.Logger) *fiber.App {
	errorHandler := func(c fiber.Ctx, err error) error {
		return handleError(c, err, log)
	}

	app := fiber.New(fiber.Config{
		AppName:         config.GetString("app.name"),
		StrictRouting:   true,
		CaseSensitive:   true,
		StructValidator: validate,
		ErrorHandler:    errorHandler,
	})

	// Register global middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	return app
}

// handleError is the global error handler for the Fiber application.
func handleError(c fiber.Ctx, err error, log *logrus.Logger) error {
	// Fiber's own errors (404, 405, etc.)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return response.ErrorResponse(c, fiberErr.Code, fiberErr.Message)
	}

	// Bind errors (malformed JSON, wrong types)
	var bindErr *fiber.BindError
	if errors.As(err, &bindErr) {
		return response.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	// Validation errors - show details
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		details := make([]fiber.Map, 0, len(valErrs))
		for _, fe := range valErrs {
			details = append(details, fiber.Map{
				"field":   fe.Field(),
				"tag":     fe.Tag(),
				"message": formatValidationError(fe),
			})
		}
		return response.ErrorResponseWithDetails(c, fiber.StatusUnprocessableEntity, "validation failed", details)
	}

	// Database errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.ErrorResponse(c, fiber.StatusNotFound, "resource not found")
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return response.ErrorResponse(c, fiber.StatusConflict, "duplicate entry")
	}

	// Foreign key violation (PostgreSQL SQLSTATE 23503)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return response.ErrorResponse(c, fiber.StatusConflict, "referenced resource does not exist")
	}

	// AppError - general business logic errors
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		if appErr.Details != nil {
			return response.ErrorResponseWithDetails(c, appErr.Code, appErr.Message, appErr.Details)
		}
		return response.ErrorResponse(c, appErr.Code, appErr.Message)
	}

	// Everything else — log the real error with logrus, return generic message
	log.WithField("error", err.Error()).WithField("path", c.Path()).Error("unhandled error")
	return response.ErrorResponse(c, fiber.StatusInternalServerError, "internal server error")
}
