package response

import "github.com/gofiber/fiber/v3"

// Response is the standard response format for all API responses.
type Response struct {
	Success bool        `json:"success"`         // Indicates if the request was successful
	Message string      `json:"message"`         // Human-readable message
	Data    interface{} `json:"data,omitempty"`  // Response data (omitted if empty)
}

// PaginationResponse is the response format for paginated data.
type PaginationResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains metadata for paginated responses.
type Pagination struct {
	Page  int `json:"page"`   // Current page number
	Limit int `json:"limit"`  // Number of items per page
	Total int `json:"total"`  // Total number of items
}

// SuccessResponse creates a successful response with status 200.
func SuccessResponse(c fiber.Ctx, message string, data interface{}) error {
	return c.JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse creates a successful response with status 201.
func CreatedResponse(c fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// PaginatedResponse creates a successful response with pagination metadata.
func PaginatedResponse(c fiber.Ctx, message string, data interface{}, page, limit, total int) error {
	return c.JSON(PaginationResponse{
		Success: true,
		Message: message,
		Data:    data,
		Pagination: Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// ErrorResponse creates an error response without data field.
func ErrorResponse(c fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(Response{
		Success: false,
		Message: message,
	})
}

// ErrorResponseWithDetails creates an error response with data field.
// Used specifically for validation errors.
func ErrorResponseWithDetails(c fiber.Ctx, code int, message string, details interface{}) error {
	return c.Status(code).JSON(Response{
		Success: false,
		Message: message,
		Data:    details,
	})
}
