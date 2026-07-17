package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/handlers"
	"github.com/ncwjsp/saep-pos/internal/menu"
	"github.com/ncwjsp/saep-pos/internal/orders"
)

func main() {
	menuStore := menu.NewDemoStore()
	orderStore := orders.NewStore(menuStore)

	router := gin.Default()
	router.GET("/t/:qrToken/menu", handlers.GetMenu(menuStore))
	router.POST("/t/:qrToken/orders", handlers.CreateOrder(orderStore))

	if err := router.Run("localhost:8080"); err != nil {
		log.Fatalf("running server: %v", err)
	}
}
