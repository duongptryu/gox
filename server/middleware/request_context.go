package middleware

import (
	pkgContext "github.com/duongptryu/gox/context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Generate/extract Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = pkgContext.WithRequestID(ctx, requestID)

		// Generate/extract Operation ID
		operationID := c.GetHeader("X-Operation-ID")
		if operationID == "" {
			operationID = uuid.New().String()
		}
		ctx = pkgContext.WithOperationID(ctx, operationID)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Add to response headers
		c.Header("X-Request-ID", requestID)
		c.Header("X-Operation-ID", operationID)

		c.Next()
	}
}
