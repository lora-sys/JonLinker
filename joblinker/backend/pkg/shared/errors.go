package shared

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeValidation  ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden   ErrorCode = "FORBIDDEN"
	ErrCodeConflict    ErrorCode = "CONFLICT"
	ErrCodeInternal    ErrorCode = "INTERNAL_ERROR"
	ErrCodeRateLimited ErrorCode = "RATE_LIMITED"
)

type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeValidation:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func NewValidationError(message string, details map[string]interface{}) *AppError {
	return &AppError{Code: ErrCodeValidation, Message: message, Details: details}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{Code: ErrCodeNotFound, Message: message}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{Code: ErrCodeUnauthorized, Message: message}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{Code: ErrCodeForbidden, Message: message}
}

func NewConflictError(message string) *AppError {
	return &AppError{Code: ErrCodeConflict, Message: message}
}

func NewInternalError(message string) *AppError {
	return &AppError{Code: ErrCodeInternal, Message: message}
}

func NewRateLimitedError(message string) *AppError {
	return &AppError{Code: ErrCodeRateLimited, Message: message}
}
