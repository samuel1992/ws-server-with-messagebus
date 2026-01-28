package services

import (
	"context"
	"log"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

type EchoService struct {
	bus        messagebus.MessageBus
	readTopic  string
	writeTopic string
	cancel     context.CancelFunc

	stopped chan struct{}
}

func NewEchoService(mb messagebus.MessageBus, readTopic, writeTopic string) Service {
	return &EchoService{
		bus:        mb,
		readTopic:  readTopic,
		writeTopic: writeTopic,
		stopped:    make(chan struct{}),
	}
}

func (s *EchoService) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	subscription := s.bus.Subscribe(s.readTopic)

	// run the service in a go routine
	go func() {
		defer func() {
			s.bus.Unsubscribe(s.readTopic, subscription)
			close(s.stopped)
		}()

		for {
			select {
			case msg := <-subscription:
				// publish the same received message to the write topic (echo)
				s.bus.Publish(s.writeTopic, msg)
			case <-ctx.Done():
				log.Println("EchoService stopping")
				return
			}
		}
	}()

	return nil
}

func (s *EchoService) Stop() error {
	log.Println("Stopping EchoService")
	if s.cancel != nil {
		s.cancel()
	}
	// Wait for goroutine to finish
	<-s.stopped

	return nil
}
