package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/handlers"
	"github.com/ncwjsp/saep-pos/internal/menu"
)

func main() {
	menuStore := menu.NewDemoStore()

	router := gin.Default()
	router.GET("/t/:qrToken/menu", handlers.GetMenu(menuStore))

	if err := router.Run("localhost:8080"); err != nil {
		log.Fatalf("running server: %v", err)
	}
}
