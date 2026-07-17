// Package menu holds the restaurant menu. Current phase: hardcoded
// single-restaurant menu in memory; a real schema replaces this later.
package menu

import "sync"

// Item is a single menu item. Prices are integers in satang (฿60.00 = 6000).
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameEn      string `json:"name_en"`
	PriceSatang int64  `json:"price_satang"`
	Category    string `json:"category"`
}

// Store serves menu items. The mutex is unused while the menu is read-only,
// but the store follows the current-phase convention (slice + sync.Mutex)
// so menu CRUD can be added without changing callers.
type Store struct {
	mu    sync.Mutex
	items []Item
}

// NewDemoStore returns a Store seeded with the hardcoded demo menu.
func NewDemoStore() *Store {
	return &Store{
		items: []Item{
			{ID: "1", Name: "ผัดกะเพราหมูสับ", NameEn: "Stir-fried basil with minced pork", PriceSatang: 6000, Category: "จานหลัก"},
			{ID: "2", Name: "ข้าวผัดกุ้ง", NameEn: "Shrimp fried rice", PriceSatang: 7000, Category: "จานหลัก"},
			{ID: "3", Name: "ผัดไทยกุ้งสด", NameEn: "Pad thai with shrimp", PriceSatang: 8000, Category: "จานหลัก"},
			{ID: "4", Name: "ต้มยำกุ้ง", NameEn: "Tom yum goong", PriceSatang: 12000, Category: "ต้ม/แกง"},
			{ID: "5", Name: "แกงเขียวหวานไก่", NameEn: "Green curry with chicken", PriceSatang: 9000, Category: "ต้ม/แกง"},
			{ID: "6", Name: "ส้มตำไทย", NameEn: "Papaya salad", PriceSatang: 5500, Category: "ยำ/สลัด"},
			{ID: "7", Name: "ข้าวเหนียวมะม่วง", NameEn: "Mango sticky rice", PriceSatang: 8500, Category: "ของหวาน"},
			{ID: "8", Name: "ชาไทยเย็น", NameEn: "Thai iced tea", PriceSatang: 4000, Category: "เครื่องดื่ม"},
			{ID: "9", Name: "น้ำเปล่า", NameEn: "Water", PriceSatang: 1500, Category: "เครื่องดื่ม"},
		},
	}
}

// List returns a copy of all menu items.
func (s *Store) List() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Item, len(s.items))
	copy(out, s.items)
	return out
}
