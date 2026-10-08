package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FiberValidator wraps validator.Validate to satisfy fiber.StructValidator.
type FiberValidator struct {
	validator *validator.Validate
}

// Validate implements fiber.StructValidator.
func (v *FiberValidator) Validate(out any) error {
	return v.validator.Struct(out)
}

// NewValidation creates a new FiberValidator instance.
func NewValidation() *FiberValidator {
	return &FiberValidator{validator: validator.New()}
}

// formatValidationError converts a validator.FieldError into a human-readable message.
func formatValidationError(fe validator.FieldError) string {
	field := fe.Field()   // Field name (e.g., "User.Email" for nested structs)
	tag := fe.Tag()       // Failed validation tag (e.g., "required", "min")
	param := fe.Param()   // Tag parameter (e.g., "3" for min=3)

	// Format message based on validation tag
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fe.Error()
	}
}
