package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ErrorCode string

const (
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeRateLimited   ErrorCode = "RATE_LIMITED"
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error         string    `json:"error"`
	Code          ErrorCode `json:"code"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			requestID, _ := c.Get("requestID")
			correlationID, _ := requestID.(string)
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			err := c.Errors.Last()
			statusCode := http.StatusInternalServerError
			code := ErrCodeInternalError

			if appErr, ok := err.Err.(*AppError); ok {
				statusCode = appErr.StatusCode
				code = appErr.Code
			}

			c.JSON(statusCode, ErrorResponse{
				Error:         err.Error(),
				Code:          code,
				CorrelationID: correlationID,
			})
		}
	}
}

type AppError struct {
	StatusCode int
	Code       ErrorCode
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(statusCode int, code ErrorCode, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func LogError(c *gin.Context, err error, userID, matchID string) {
	requestID, _ := c.Get("requestID")
	correlationID, _ := requestID.(string)
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	log.Printf("ERROR: %s | user_id=%s | match_id=%s | correlation_id=%s | type=%T",
		err.Error(), userID, matchID, correlationID, err)
}
