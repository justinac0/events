package internal

import "sync"

type EventBus struct {
	subscribers map[string][]chan interface{}
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan interface{}),
	}
}

func (eb *EventBus) Subscribe(topic string, ch chan interface{}) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[topic] = append(eb.subscribers[topic], ch)
}

func (eb *EventBus) Publish(topic string, data interface{}) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, ch := range eb.subscribers[topic] {
		go func(ch chan interface{}) {
			ch <- data
		}(ch)
	}
}
