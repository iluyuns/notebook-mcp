package oauth

import (
	"net/http"
	"strings"

	"notebook-mcp/internal/authctx"

	"github.com/gin-gonic/gin"
)

func BearerAuthMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
			return
		}
		token := strings.TrimSpace(parts[1])
		userID, ok := svc.UserIDFromAccessToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}
		c.Request = c.Request.WithContext(authctx.WithUserID(c.Request.Context(), userID))
		c.Next()
	}
}
