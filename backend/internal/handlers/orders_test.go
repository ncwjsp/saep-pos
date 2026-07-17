package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/menu"
	"github.com/ncwjsp/saep-pos/internal/orders"
	"github.com/ncwjsp/saep-pos/internal/sse"
)

func newOrderTestRouter() (*gin.Engine, *orders.Store, *sse.Hub) {
	gin.SetMode(gin.TestMode)
	store := orders.NewStore(menu.NewDemoStore())
	hub := sse.NewHub()
	r := gin.New()
	r.POST("/t/:qrToken/orders", CreateOrder(store, hub))
	return r, store, hub
}

func postOrder(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestCreateOrder(t *testing.T) {
	r, store, _ := newOrderTestRouter()
	body := `{"items":[{"menu_item_id":"1","quantity":2,"note":"ไม่เผ็ด"},{"menu_item_id":"8","quantity":1}]}`
	w := postOrder(t, r, "/t/demo/orders", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusCreated, w.Body.String())
	}

	var o orders.Order
	if err := json.Unmarshal(w.Body.Bytes(), &o); err != nil {
		t.Fatalf("unmarshaling response: %v", err)
	}
	if o.Status != orders.StatusPending {
		t.Errorf("status = %q, want %q", o.Status, orders.StatusPending)
	}
	if len(o.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(o.Items))
	}
	// Price snapshotted from the menu: 2×6000 (กะเพรา) + 1×4000 (ชาไทย).
	if o.Items[0].PriceSatang != 6000 {
		t.Errorf("items[0].price_satang = %d, want 6000", o.Items[0].PriceSatang)
	}
	if o.Items[0].Name != "ผัดกะเพราหมูสับ" {
		t.Errorf("items[0].name = %q, want Thai name snapshot", o.Items[0].Name)
	}
	if o.Items[0].Note != "ไม่เผ็ด" {
		t.Errorf("items[0].note = %q, want ไม่เผ็ด", o.Items[0].Note)
	}
	if o.TotalSatang != 16000 {
		t.Errorf("total_satang = %d, want 16000", o.TotalSatang)
	}
	if o.CreatedAt.IsZero() {
		t.Error("created_at is zero")
	}

	if got := store.List(); len(got) != 1 {
		t.Errorf("stored orders = %d, want 1", len(got))
	}
}

func TestCreateOrderPublishesEvent(t *testing.T) {
	r, _, hub := newOrderTestRouter()
	events, cancel := hub.Subscribe()
	defer cancel()

	w := postOrder(t, r, "/t/demo/orders", `{"items":[{"menu_item_id":"1","quantity":1}]}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	select {
	case ev := <-events:
		if ev.Name != "order_created" {
			t.Errorf("event name = %q, want order_created", ev.Name)
		}
		var o orders.Order
		if err := json.Unmarshal(ev.Data, &o); err != nil {
			t.Fatalf("unmarshaling event data: %v", err)
		}
		if o.TotalSatang != 6000 {
			t.Errorf("event order total_satang = %d, want 6000", o.TotalSatang)
		}
	default:
		t.Fatal("no event published on order creation")
	}
}

func TestCreateOrderValidation(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
		want int
	}{
		{"unknown token", "/t/nope/orders", `{"items":[{"menu_item_id":"1","quantity":1}]}`, http.StatusNotFound},
		{"malformed json", "/t/demo/orders", `{not json`, http.StatusBadRequest},
		{"empty items", "/t/demo/orders", `{"items":[]}`, http.StatusBadRequest},
		{"missing items", "/t/demo/orders", `{}`, http.StatusBadRequest},
		{"zero quantity", "/t/demo/orders", `{"items":[{"menu_item_id":"1","quantity":0}]}`, http.StatusBadRequest},
		{"negative quantity", "/t/demo/orders", `{"items":[{"menu_item_id":"1","quantity":-3}]}`, http.StatusBadRequest},
		{"excessive quantity", "/t/demo/orders", `{"items":[{"menu_item_id":"1","quantity":999}]}`, http.StatusBadRequest},
		{"unknown menu item", "/t/demo/orders", `{"items":[{"menu_item_id":"999","quantity":1}]}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, store, _ := newOrderTestRouter()
			w := postOrder(t, r, tt.path, tt.body)
			if w.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.want, w.Body.String())
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshaling response: %v", err)
			}
			if body["error"] == "" {
				t.Error("response missing error field")
			}
			if got := store.List(); len(got) != 0 {
				t.Errorf("stored orders = %d, want 0", len(got))
			}
		})
	}
}
