// Package sse is a fan-out hub for server-sent events: the kitchen stream
// subscribes, order handlers publish.
package sse

import "sync"

// Event is a named payload sent to subscribers. Data is pre-marshaled JSON;
// the hub never encodes or inspects it.
type Event struct {
	Name string
	Data []byte
}

// subscriberBuffer absorbs short bursts per subscriber. If a subscriber
// falls further behind than this, events are dropped for it (see Publish).
const subscriberBuffer = 16

// Hub fans published events out to all current subscribers.
// The zero value is not usable; call NewHub.
type Hub struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan Event]struct{})}
}

// Subscribe registers a subscriber and returns its event channel plus a
// cancel func. Cancel is idempotent and closes the channel; the caller must
// not close it. After cancel, the subscriber receives nothing further.
func (h *Hub) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, subscriberBuffer)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subs, ch)
			h.mu.Unlock()
			// Safe to close only now: the channel is out of the map, and
			// Publish sends exclusively to channels in the map under mu.
			close(ch)
		})
	}
	return ch, cancel
}

// Publish sends ev to every subscriber. Sends are non-blocking: a subscriber
// whose buffer is full misses this event rather than stalling the hub and
// every other subscriber behind it.
func (h *Hub) Publish(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}
