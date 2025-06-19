package middleware

import (
	"errors"
	"net/http"

	"github.com/duongptryu/gox/logger"
	"github.com/duongptryu/gox/response"
	"github.com/duongptryu/gox/syserr"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

func handleError(c *gin.Context, err error) {
	var sysErr *syserr.Error
	if errors.As(err, &sysErr) {
		// statusCode := getHTTPStatusCode(sysErr.Code())
		c.JSON(http.StatusOK, response.NewErrorResponse(
			string(sysErr.Code()),
			sysErr.Error(),
			nil,
		))
		return
	}

	// log error
	logger.LogError(c.Request.Context(), err)

	// Default error
	c.JSON(http.StatusOK, response.NewErrorResponse(
		"internal_error",
		"An error occurred",
		nil,
	))
}
