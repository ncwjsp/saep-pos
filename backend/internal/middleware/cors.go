// Package middleware holds Gin middleware shared across routes.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// devOrigins are the frontend dev servers allowed to call the API.
// Deployed origins are configured in a later phase.
var devOrigins = map[string]bool{
	"http://localhost:3000": true,
	"http://127.0.0.1:3000": true,
}

// CORS allows cross-origin requests from the dev frontend.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin := c.GetHeader("Origin"); devOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
