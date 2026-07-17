package handlers

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/sse"
)

func TestListOrders(t *testing.T) {
	r, store, _ := newOrderTestRouter()
	r.GET("/kitchen/orders", ListOrders(store))

	postOrder(t, r, "/t/demo/orders", `{"items":[{"menu_item_id":"1","quantity":1}]}`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/kitchen/orders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"orders":[`) {
		t.Errorf("body missing orders array: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"total_satang":6000`) {
		t.Errorf("body missing created order: %s", w.Body.String())
	}
}

func TestKitchenStreamDeliversEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := sse.NewHub()
	r := gin.New()
	r.GET("/kitchen/stream", KitchenStream(hub))
	srv := httptest.NewServer(r)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/kitchen/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}

	// The client may connect before the handler has subscribed, so publish
	// repeatedly until the event arrives.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		tick := time.NewTicker(50 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
				hub.Publish(sse.Event{Name: "order_created", Data: []byte(`{"id":"42"}`)})
			}
		}
	}()

	scanner := bufio.NewScanner(resp.Body)
	var sawEvent bool
	for scanner.Scan() {
		line := scanner.Text()
		if line == "event: order_created" {
			sawEvent = true
		}
		if sawEvent && line == `data: {"id":"42"}` {
			return // full event received
		}
	}
	t.Fatalf("stream ended without delivering the event (scanner err: %v)", scanner.Err())
}
