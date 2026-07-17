package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/orders"
	"github.com/ncwjsp/saep-pos/internal/sse"
)

type createOrderRequest struct {
	Items []struct {
		MenuItemID string `json:"menu_item_id"`
		Quantity   int    `json:"quantity"`
		Note       string `json:"note"`
	} `json:"items"`
}

// CreateOrder submits an order for the table identified by :qrToken and
// publishes an order_created event for the kitchen stream.
func CreateOrder(store *orders.Store, hub *sse.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Param("qrToken") != demoQRToken {
			c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
			return
		}

		var req createOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		items := make([]orders.NewItem, 0, len(req.Items))
		for _, it := range req.Items {
			items = append(items, orders.NewItem{
				MenuItemID: it.MenuItemID,
				Quantity:   it.Quantity,
				Note:       it.Note,
			})
		}

		o, err := store.Create(items)
		if err != nil {
			if errors.Is(err, orders.ErrEmptyOrder) ||
				errors.Is(err, orders.ErrInvalidQuantity) ||
				errors.Is(err, orders.ErrUnknownMenuItem) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "creating order failed"})
			return
		}

		if data, err := json.Marshal(o); err == nil {
			hub.Publish(sse.Event{Name: "order_created", Data: data})
		}

		c.JSON(http.StatusCreated, o)
	}
}
