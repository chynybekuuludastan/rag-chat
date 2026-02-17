package model

import "fmt"

type AppError struct {
	Code    int    `json:"-"`
	Err     string `json:"error"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    422,
		Err:     "validation_error",
		Message: message,
	}
}

func NewAppError(code int, err, message string) *AppError {
	return &AppError{
		Code:    code,
		Err:     err,
		Message: message,
	}
}

func WrapInternal(err error) *AppError {
	return &AppError{
		Code:    500,
		Err:     "internal_error",
		Message: fmt.Sprintf("internal error: %v", err),
	}
}

var (
	ErrNotFound     = &AppError{Code: 404, Err: "not_found", Message: "Resource not found"}
	ErrUnauthorized = &AppError{Code: 401, Err: "unauthorized", Message: "Invalid or expired token"}
	ErrForbidden    = &AppError{Code: 403, Err: "forbidden", Message: "Access denied"}
	ErrConflict     = &AppError{Code: 409, Err: "conflict", Message: "Resource already exists"}
	ErrRateLimit    = &AppError{Code: 429, Err: "rate_limit_exceeded", Message: "Too many requests"}
	ErrInternal     = &AppError{Code: 500, Err: "internal_error", Message: "Something went wrong"}
)
