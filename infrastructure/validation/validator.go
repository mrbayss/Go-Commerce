package validation

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

// FiberValidator wraps validator.Validate to satisfy fiber.StructValidator.
type FiberValidator struct {
	Validator *validator.Validate
}

// Validate implements fiber.StructValidator.
func (v *FiberValidator) Validate(out any) error {
	return v.Validator.Struct(out)
}

// NewValidation creates a new FiberValidator instance.
func NewValidation() *FiberValidator {
	validate := validator.New()

	_ = validate.RegisterValidation("id_phone", validateIDPhone)
	_ = validate.RegisterValidation("password", validatePassword)
	_ = validate.RegisterValidation("notblank", validators.NotBlank)

	return &FiberValidator{Validator: validate}
}

// formatValidationError converts a validator.FieldError into a human-readable message.
func FormatValidationError(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "notblank":
		return fmt.Sprintf("%s must not be empty or whitespace only", field)
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
	case "id_phone":
		return fmt.Sprintf("%s must be a valid Indonesian phone number starting with 62", field)
	case "password":
		return fmt.Sprintf("%s must contain uppercase, lowercase, numbers, symbols, and no spaces", field)
	default:
		return fe.Error()
	}
}

// validation id phone number
var idPhoneRegex = regexp.MustCompile(`^628[1-9][0-9]{7,11}$`)

func validateIDPhone(fl validator.FieldLevel) bool {
	return idPhoneRegex.MatchString(fl.Field().String())
}

// validatePassword:
// 1. Uppercase and lowercase letters
// 2. Numbers
// 3. Symbols
// 4. No whitespace
func validatePassword(fl validator.FieldLevel) bool {
	pass := fl.Field().String()

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
		hasSymbol bool
	)

	for _, r := range pass {
		if unicode.IsSpace(r) {
			return false
		}
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSymbol
}
