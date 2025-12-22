package services

import (
	"context"
	"testing"
	"time"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

func TestTimeNowServicePublishesTime(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "time:from-ws"
	writeTopic := "time:to-ws"

	service := NewTimeNowService(bus, readTopic, writeTopic)
	service.Start(context.Background())

	outputCh := bus.Subscribe(writeTopic)

	select {
	case msg := <-outputCh:
		_, err := time.Parse(time.RFC3339, string(msg))
		if err != nil {
			t.Errorf("invalid RFC3339 format: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for time message")
	}

	service.Stop()
}

func TestTimeNowServiceMultiplePublishes(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "time:from-ws"
	writeTopic := "time:to-ws"

	service := NewTimeNowService(bus, readTopic, writeTopic)
	service.Start(context.Background())

	outputCh := bus.Subscribe(writeTopic)

	for i := 0; i < 3; i++ {
		select {
		case msg := <-outputCh:
			parsed, err := time.Parse(time.RFC3339, string(msg))
			if err != nil {
				t.Errorf("message %d: invalid RFC3339 format: %v", i, err)
			}
			if time.Since(parsed) > 5*time.Second {
				t.Errorf("message %d: time too old: %s", i, parsed)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("message %d: timeout waiting for time message", i)
		}
	}

	service.Stop()
}

func TestTimeNowServiceStop(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "time:from-ws"
	writeTopic := "time:to-ws"

	service := NewTimeNowService(bus, readTopic, writeTopic)
	service.Start(context.Background())

	outputCh := bus.Subscribe(writeTopic)

	select {
	case <-outputCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for first message")
	}

	service.Stop()

	time.Sleep(3 * time.Second)

	messagesAfterStop := 0
	for len(outputCh) > 0 {
		<-outputCh
		messagesAfterStop++
	}

	if messagesAfterStop > 1 {
		t.Errorf("received %d messages after stop", messagesAfterStop)
	}
}

func TestTimeNowServiceContextCancellation(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "time:from-ws"
	writeTopic := "time:to-ws"

	service := NewTimeNowService(bus, readTopic, writeTopic)
	ctx, cancel := context.WithCancel(context.Background())

	service.Start(ctx)

	outputCh := bus.Subscribe(writeTopic)

	select {
	case <-outputCh:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	cancel()

	time.Sleep(3 * time.Second)

	messagesAfterCancel := 0
	for len(outputCh) > 0 {
		<-outputCh
		messagesAfterCancel++
	}

	if messagesAfterCancel > 1 {
		t.Errorf("received %d messages after context cancel", messagesAfterCancel)
	}
}

func TestTimeNowServiceStopBlocks(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "time:from-ws"
	writeTopic := "time:to-ws"

	service := NewTimeNowService(bus, readTopic, writeTopic)
	service.Start(context.Background())

	done := make(chan struct{})
	go func() {
		service.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Stop() did not complete in time")
	}
}
