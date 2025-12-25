package services

import (
	"context"
	"log"
	"time"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

type TimeNowService struct {
	bus        messagebus.MessageBus
	readTopic  string
	writeTopic string
	cancel     context.CancelFunc

	stopped chan struct{}
}

func NewTimeNowService(mb messagebus.MessageBus, readTopic, writeTopic string) Service {
	return &TimeNowService{
		bus:        mb,
		readTopic:  readTopic,
		writeTopic: writeTopic,
		stopped:    make(chan struct{}),
	}
}

func (s *TimeNowService) Start(c context.Context) error {
	ctx, cancel := context.WithCancel(c)
	s.cancel = cancel

	ticker := time.NewTicker(2 * time.Second)

	go func() {

		defer func() {
			close(s.stopped)
		}()

		for {
			select {
			case <-ticker.C:
				datetime := time.Now().Format(time.RFC3339)
				s.bus.Publish(s.writeTopic, []byte(datetime))
			case <-ctx.Done():
				log.Println("TimeNowService stopping")
				return
			}
		}

	}()

	return nil
}

func (s *TimeNowService) Stop() error {
	log.Println("Stopping TimeNowService")
	if s.cancel != nil {
		s.cancel()
	}

	<-s.stopped

	return nil
}
