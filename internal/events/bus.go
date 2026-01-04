package events

import "sync"

type Event struct {
	Type string
	Data any
}

type EventBus struct {
	subscribers []chan Event
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make([]chan Event, 0),
	}
}

func (eb *EventBus) Subscribe() chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 100)
	eb.subscribers = append(eb.subscribers, ch)
	return ch
}

func (eb *EventBus) Unsubscribe(ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for i, sub := range eb.subscribers {
		if sub == ch {
			close(ch)
			eb.subscribers = append(eb.subscribers[:i], eb.subscribers[i+1:]...)
			break
		}
	}
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, sub := range eb.subscribers {
		select {
		case sub <- event:
		default:
			// Skip if channel is full
		}
	}
}
