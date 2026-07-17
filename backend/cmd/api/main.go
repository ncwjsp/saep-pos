package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/handlers"
	"github.com/ncwjsp/saep-pos/internal/menu"
	"github.com/ncwjsp/saep-pos/internal/middleware"
	"github.com/ncwjsp/saep-pos/internal/orders"
	"github.com/ncwjsp/saep-pos/internal/sse"
)

func main() {
	menuStore := menu.NewDemoStore()
	orderStore := orders.NewStore(menuStore)
	hub := sse.NewHub()

	router := gin.Default()
	router.Use(middleware.CORS())
	router.GET("/t/:qrToken/menu", handlers.GetMenu(menuStore))
	router.POST("/t/:qrToken/orders", handlers.CreateOrder(orderStore, hub))
	router.GET("/kitchen/orders", handlers.ListOrders(orderStore))
	router.GET("/kitchen/stream", handlers.KitchenStream(hub))

	// Bind all interfaces so phones on the same LAN can reach the API
	// (the customer page is opened from a phone in the QR demo).
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("running server: %v", err)
	}
}
