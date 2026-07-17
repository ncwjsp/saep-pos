package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/menu"
)

// demoQRToken is the only valid table token in the current phase.
const demoQRToken = "demo"

// GetMenu returns the menu for the table identified by :qrToken.
func GetMenu(store *menu.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Param("qrToken") != demoQRToken {
			c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": store.List()})
	}
}
