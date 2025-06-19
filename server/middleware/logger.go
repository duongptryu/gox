package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// use this when to want to customize the logger std output
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format(time.DateTime),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	})
}
