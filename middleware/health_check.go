package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Configures health check handler.
// Returns handler configured for health check at path.
// A successful health check returns status 200 and JSON `{"ok": true"}`.
func HealthCheck(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == path {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"ok": true})
		} else {
			c.Next()
		}
	}
}
