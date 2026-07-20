package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type EventType string

const (
	EventUserRegistered  EventType = "user.registered"
	EventPasswordReset   EventType = "user.password_reset"
	EventEmailVerified   EventType = "user.email_verified"
	EventMessageReceived EventType = "message.received"
)

type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

func NewEvent(eventType EventType, payload any) (Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}
	return Event{
		ID:        fmt.Sprintf("%s-%d", eventType, time.Now().UnixNano()),
		Type:      eventType,
		Payload:   data,
		Timestamp: time.Now().UTC(),
	}, nil
}

func (e Event) PayloadInto(target any) error {
	return json.Unmarshal(e.Payload, target)
}

type EventBus struct {
	subscribers map[EventType][]chan Event
	mu          sync.RWMutex
	bufferSize  int
	closed      bool // track if shutdown was called
}

func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
		bufferSize:  bufferSize,
	}
}

func (eb *EventBus) Subscribe(eventTypes ...EventType) <-chan Event {
	ch := make(chan Event, eb.bufferSize)

	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		close(ch) // immediately close if already shut down
		return ch
	}

	for _, et := range eventTypes {
		eb.subscribers[et] = append(eb.subscribers[et], ch)
	}

	return ch
}

func (eb *EventBus) Unsubscribe(ch <-chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Don't close if already shut down — Shutdown() already did it
	if eb.closed {
		return
	}

	for et, subs := range eb.subscribers {
		for i, sub := range subs {
			if sub == ch {
				// Only close if we haven't already
				select {
				case <-sub:
					// already closed/empty, skip
				default:
					close(sub)
				}
				eb.subscribers[et] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	eb.mu.RLock()
	if eb.closed {
		eb.mu.RUnlock()
		return fmt.Errorf("event bus is shut down")
	}
	subs := make([]chan Event, len(eb.subscribers[event.Type]))
	copy(subs, eb.subscribers[event.Type])
	eb.mu.RUnlock()

	if len(subs) == 0 {
		return nil
	}

	for _, ch := range subs {
		select {
		case ch <- event:
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			continue
		}
	}

	return nil
}

func (eb *EventBus) Shutdown() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return // idempotent — safe to call multiple times
	}

	eb.closed = true

	for _, subs := range eb.subscribers {
		for _, ch := range subs {
			// Safe close: check if already closed via recover
			func(c chan Event) {
				defer func() { recover() }() // catch "close of closed channel"
				close(c)
			}(ch)
		}
	}
	eb.subscribers = make(map[EventType][]chan Event)
}
