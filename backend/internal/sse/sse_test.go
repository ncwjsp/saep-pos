package sse

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func recvOrTimeout(t *testing.T, ch <-chan Event) Event {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
		return Event{}
	}
}

func TestPublishFansOutToAllSubscribers(t *testing.T) {
	h := NewHub()
	var chans []<-chan Event
	for i := 0; i < 3; i++ {
		ch, cancel := h.Subscribe()
		defer cancel()
		chans = append(chans, ch)
	}

	h.Publish(Event{Name: "order_created", Data: []byte(`{"id":"1"}`)})

	for i, ch := range chans {
		ev := recvOrTimeout(t, ch)
		if ev.Name != "order_created" {
			t.Errorf("subscriber %d got event %q, want order_created", i, ev.Name)
		}
	}
}

func TestCancelledSubscriberGetsNothing(t *testing.T) {
	h := NewHub()
	ch, cancel := h.Subscribe()
	cancel()

	h.Publish(Event{Name: "order_created"}) // must not panic (send on closed chan)

	if ev, ok := <-ch; ok {
		t.Errorf("received %q on cancelled subscription, want closed channel", ev.Name)
	}
}

func TestCancelIsIdempotent(t *testing.T) {
	h := NewHub()
	_, cancel := h.Subscribe()
	cancel()
	cancel() // must not panic (double close)
}

func TestSlowSubscriberDoesNotBlockOthers(t *testing.T) {
	h := NewHub()
	slow, cancelSlow := h.Subscribe()
	defer cancelSlow()
	_ = slow // never read: its buffer fills up

	fast, cancelFast := h.Subscribe()
	defer cancelFast()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < subscriberBuffer*2; i++ {
			h.Publish(Event{Name: "e"})
			recvOrTimeout(t, fast)
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publishing blocked on a slow subscriber")
	}
}

func TestConcurrentSubscribePublishCancel(t *testing.T) {
	// Exercises the hub from many goroutines at once; run under -race
	// (WSL/CI on this machine) to get the race detector's verdict.
	h := NewHub()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ch, cancel := h.Subscribe()
			h.Publish(Event{Name: fmt.Sprintf("e%d", i)})
			select {
			case <-ch:
			default:
			}
			cancel()
		}(i)
	}
	wg.Wait()
}
