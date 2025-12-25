package services

import (
	"context"
	"log"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

type PingService struct {
	bus        messagebus.MessageBus
	readTopic  string
	writeTopic string
	cancel     context.CancelFunc

	stopped chan struct{}
}

func NewPingService(mb messagebus.MessageBus, readTopic, writeTopic string) Service {
	return &PingService{
		bus:        mb,
		readTopic:  readTopic,
		writeTopic: writeTopic,
		stopped:    make(chan struct{}),
	}
}

func (s *PingService) Start(ctx context.Context) error {
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
				s.bus.Publish(s.writeTopic, msg)
			case <-ctx.Done():
				log.Println("PingService stopping")
				return
			}
		}
	}()

	return nil
}

func (s *PingService) Stop() error {
	log.Println("Stopping PingService")
	if s.cancel != nil {
		s.cancel()
	}
	// Wait for goroutine to finish
	<-s.stopped

	return nil
}
