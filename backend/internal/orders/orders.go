// Package orders holds customer orders. Current phase: in-memory storage
// for the hardcoded demo table; sessions and the DB transaction come later.
package orders

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ncwjsp/saep-pos/internal/menu"
)

// Status is an order's kitchen status: pending → preparing → served (or cancelled).
type Status string

const (
	StatusPending   Status = "pending"
	StatusPreparing Status = "preparing"
	StatusServed    Status = "served"
	StatusCancelled Status = "cancelled"
)

// maxQuantity bounds a single line item so a typo can't order 9999 dishes.
const maxQuantity = 50

var (
	ErrEmptyOrder      = errors.New("order has no items")
	ErrInvalidQuantity = errors.New("quantity must be between 1 and 50")
	ErrUnknownMenuItem = errors.New("unknown menu item")
)

// Item is an order line with the menu price snapshotted at order time.
type Item struct {
	MenuItemID  string `json:"menu_item_id"`
	Name        string `json:"name"`
	PriceSatang int64  `json:"price_satang"`
	Quantity    int    `json:"quantity"`
	Note        string `json:"note,omitempty"`
}

// Order is a submitted customer order.
type Order struct {
	ID          string    `json:"id"`
	Status      Status    `json:"status"`
	Items       []Item    `json:"items"`
	TotalSatang int64     `json:"total_satang"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewItem is a requested line item, before validation and price snapshot.
type NewItem struct {
	MenuItemID string
	Quantity   int
	Note       string
}

// Store validates and holds orders in memory.
type Store struct {
	mu     sync.Mutex
	menu   *menu.Store
	orders []Order
	nextID int
}

func NewStore(m *menu.Store) *Store {
	return &Store{menu: m, nextID: 1}
}

// Create validates the requested items against the menu, snapshots current
// prices, and stores the order with status pending.
func (s *Store) Create(items []NewItem) (Order, error) {
	if len(items) == 0 {
		return Order{}, ErrEmptyOrder
	}

	lines := make([]Item, 0, len(items))
	var total int64
	for _, ni := range items {
		if ni.Quantity < 1 || ni.Quantity > maxQuantity {
			return Order{}, fmt.Errorf("item %q: %w", ni.MenuItemID, ErrInvalidQuantity)
		}
		mi, ok := s.menu.Get(ni.MenuItemID)
		if !ok {
			return Order{}, fmt.Errorf("item %q: %w", ni.MenuItemID, ErrUnknownMenuItem)
		}
		lines = append(lines, Item{
			MenuItemID:  mi.ID,
			Name:        mi.Name,
			PriceSatang: mi.PriceSatang,
			Quantity:    ni.Quantity,
			Note:        ni.Note,
		})
		total += mi.PriceSatang * int64(ni.Quantity)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	o := Order{
		ID:          strconv.Itoa(s.nextID),
		Status:      StatusPending,
		Items:       lines,
		TotalSatang: total,
		CreatedAt:   time.Now().UTC(),
	}
	s.nextID++
	s.orders = append(s.orders, o)
	return o, nil
}

// List returns a copy of all orders, oldest first.
func (s *Store) List() []Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Order, len(s.orders))
	copy(out, s.orders)
	return out
}
