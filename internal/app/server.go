package app

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mrbayss/Go-Commerce/internal/config"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// NewFiber creates and configures a new Fiber application instance.
func NewFiber(cfg *config.Config, log *logrus.Logger, validate *FiberValidator) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         cfg.App.Name,
		StrictRouting:   true,
		CaseSensitive:   true,
		StructValidator: validate,
		ErrorHandler:    errorHandler(log),
		BodyLimit:       10 * 1024 * 1024,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
	})

	app.Use(requestid.New())
	app.Use(func(c fiber.Ctx) error {
		ctx := context.WithValue(c.Context(), "request_id", c.GetRespHeader("X-Request-Id"))
		c.SetContext(ctx)
		c.Locals("request_id", c.GetRespHeader("X-Request-Id"))
		return c.Next()
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
	app.Use(accessLogger(log))

	return app
}

func accessLogger(log *logrus.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()

		status := c.Response().StatusCode()
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			status = fiberErr.Code
		}
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			status = appErr.Code
		}

		fields := logrus.Fields{
			"method": c.Method(),
			"path":   c.Path(),
			"status": status,
		}
		if rid, ok := c.Locals("request_id").(string); ok {
			fields["rid"] = rid
		}

		entry := log.WithFields(fields)
		if err != nil {
			entry.WithError(err).Info("request completed with error")
		} else {
			entry.Info("request completed")
		}

		return err
	}
}

func errorHandler(log *logrus.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
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
					"message": FormatValidationError(fe),
				})
			}
			return response.ErrorResponseWithDetails(c, fiber.StatusUnprocessableEntity, "validation failed", details)
		}

		// Database errors
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ErrorResponse(c, fiber.StatusNotFound, "resource not found")
		}

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			msg := "duplicate entry"
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.ConstraintName == "idx_users_email" {
					msg = "email already registered"
				} else if pgErr.ConstraintName == "idx_users_phone" {
					msg = "phone number already registered"
				}
			}
			return response.ErrorResponse(c, fiber.StatusConflict, msg)
		}

		// Foreign key violation (PostgreSQL SQLSTATE 23503)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return response.ErrorResponse(c, fiber.StatusConflict, "referenced resource does not exist")
		}

		// AppError - general business logic errors
		var appErr *apperrors.AppError
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
}
