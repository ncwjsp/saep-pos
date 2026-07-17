package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ncwjsp/saep-pos/internal/menu"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/t/:qrToken/menu", GetMenu(menu.NewDemoStore()))
	return r
}

func TestGetMenuDemo(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t/demo/menu", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body struct {
		Items []menu.Item `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshaling response: %v", err)
	}
	if len(body.Items) == 0 {
		t.Fatal("menu is empty")
	}

	first := body.Items[0]
	if first.Name != "ผัดกะเพราหมูสับ" {
		t.Errorf("first item name = %q, want Thai name intact", first.Name)
	}
	if first.PriceSatang != 6000 {
		t.Errorf("first item price_satang = %d, want 6000", first.PriceSatang)
	}
}

func TestGetMenuUnknownToken(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t/nope/menu", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshaling response: %v", err)
	}
	if body["error"] == "" {
		t.Error("response missing error field")
	}
}
