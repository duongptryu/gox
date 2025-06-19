package middleware

import (
	"strings"

	"github.com/duongptryu/gox/auth"
	"github.com/duongptryu/gox/context"
	"github.com/duongptryu/gox/syserr"

	"github.com/gin-gonic/gin"
)

// RequireAuth validates JWT tokens and sets user context
func RequireAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.Error(syserr.New(syserr.UnauthorizedCode, "authorization token required"))
			return
		}

		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			c.Error(err)
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithUserID(ctx, claims.UserID)
		ctx = context.WithUserType(ctx, claims.UserType)
		ctx = context.WithAuthClaims(ctx, claims)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func extractTokenFromHeader(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
