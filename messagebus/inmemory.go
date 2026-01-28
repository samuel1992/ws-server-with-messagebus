package messagebus

import (
	"log"
	"sync"
)

type InMemoryMessageBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan []byte
}

func NewInMemoryMessageBus() MessageBus {
	return &InMemoryMessageBus{
		subscribers: make(map[string][]chan []byte),
	}
}

func (mb *InMemoryMessageBus) Subscribe(topic string) chan []byte {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan []byte, 256)
	mb.subscribers[topic] = append(mb.subscribers[topic], ch)

	return ch
}

func (mb *InMemoryMessageBus) Unsubscribe(topic string, ch chan []byte) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	channels := mb.subscribers[topic]
	for i, c := range channels {
		if c == ch {
			mb.subscribers[topic] = append(channels[:i], channels[i+1:]...)

			// Close the channel to signal completion
			close(ch)
			break
		}
	}
}

func (mb *InMemoryMessageBus) Publish(topic string, msg []byte) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	for _, ch := range mb.subscribers[topic] {
		select {
		case ch <- msg:
		default:
			log.Printf("Warning: subscriber channel full, dropping message")
		}
	}
}
