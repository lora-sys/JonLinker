package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"joblinker/pkg/shared"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Error: %v", err)

			if appErr, ok := err.Err.(*shared.AppError); ok {
				c.JSON(appErr.HTTPStatus(), gin.H{
					"error":  appErr.Message,
					"code":   appErr.Code,
					"status": "error",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Internal server error",
				"status": "error",
			})
		}
	}
}

func LogError(c *gin.Context, err error, userID string, correlationID string) {
	log.Printf("Error: %v | User: %s | RequestID: %s", err, userID, correlationID)
}
