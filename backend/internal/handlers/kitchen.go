package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/sse"
)

// heartbeatInterval keeps idle streams alive through proxies and lets the
// server notice dead connections.
const heartbeatInterval = 20 * time.Second

// KitchenStream streams hub events to the kitchen as server-sent events.
func KitchenStream(hub *sse.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		events, cancel := hub.Subscribe()
		defer cancel()

		// Flush headers immediately so EventSource fires its open event.
		c.Writer.Flush()

		heartbeat := time.NewTicker(heartbeatInterval)
		defer heartbeat.Stop()

		ctx := c.Request.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-events:
				fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev.Name, ev.Data)
				c.Writer.Flush()
			case <-heartbeat.C:
				fmt.Fprint(c.Writer, ": ping\n\n")
				c.Writer.Flush()
			}
		}
	}
}
