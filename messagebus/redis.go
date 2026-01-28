package messagebus

import (
	"context"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisMessageBus struct {
	client        *redis.Client
	ctx           context.Context
	mu            sync.RWMutex
	subscriptions map[chan []byte]*subscription
}

type subscription struct {
	pubsub *redis.PubSub
	done   chan struct{}
}

func NewRedisMessageBus(options *redis.Options) MessageBus {
	client := redis.NewClient(options)
	return &RedisMessageBus{
		client:        client,
		ctx:           context.Background(),
		subscriptions: make(map[chan []byte]*subscription),
	}
}

func (mb *RedisMessageBus) Subscribe(topic string) chan []byte {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan []byte, 256)
	pubsub := mb.client.Subscribe(mb.ctx, topic)

	_, err := pubsub.Receive(mb.ctx)
	if err != nil {
		log.Printf("Error subscribing to topic %s: %v", topic, err)
		close(ch)
		return ch
	}

	sub := &subscription{
		pubsub: pubsub,
		done:   make(chan struct{}),
	}
	mb.subscriptions[ch] = sub

	go func() {
		defer func() {
			close(sub.done)
			close(ch)
		}()

		redisCh := pubsub.Channel()
		for msg := range redisCh {
			select {
			case ch <- []byte(msg.Payload):
			default:
				log.Printf("Warning: subscriber channel full, dropping message")
			}
		}
	}()

	return ch
}

func (mb *RedisMessageBus) Unsubscribe(topic string, ch chan []byte) {
	mb.mu.Lock()
	sub, ok := mb.subscriptions[ch]
	if !ok {
		mb.mu.Unlock()
		return
	}
	delete(mb.subscriptions, ch)
	mb.mu.Unlock()

	if err := sub.pubsub.Close(); err != nil {
		log.Printf("Error closing redis pubsub: %v", err)
	}

	<-sub.done
}

func (mb *RedisMessageBus) Publish(topic string, msg []byte) {
	err := mb.client.Publish(mb.ctx, topic, msg).Err()
	if err != nil {
		log.Printf("Error publishing message to topic %s: %v", topic, err)
	}
}
