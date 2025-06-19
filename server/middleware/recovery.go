package middleware

import (
	"net/http"

	"github.com/duongptryu/gox/logger"
	"github.com/duongptryu/gox/response"
	"github.com/duongptryu/gox/syserr"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		logger.LogError(c.Request.Context(), err.(error))

		response.NewErrorResponse(string(syserr.InternalCode), "internal server error", err).
			JSON(c, http.StatusInternalServerError)
	})
}
