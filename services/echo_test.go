package services

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

func TestEchoServiceEcho(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "echo:from-ws"
	writeTopic := "echo:to-ws"

	service := NewEchoService(bus, readTopic, writeTopic)
	ctx := context.Background()

	err := service.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	outputCh := bus.Subscribe(writeTopic)

	bus.Publish(readTopic, []byte("hello"))

	select {
	case msg := <-outputCh:
		if !bytes.Equal(msg, []byte("hello")) {
			t.Errorf("expected 'hello', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for echo")
	}

	service.Stop()
}

func TestEchoServiceMultipleMessages(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "echo:from-ws"
	writeTopic := "echo:to-ws"

	service := NewEchoService(bus, readTopic, writeTopic)
	service.Start(context.Background())

	outputCh := bus.Subscribe(writeTopic)

	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}

	for _, msg := range messages {
		bus.Publish(readTopic, []byte(msg))
	}

	for i, expected := range messages {
		select {
		case msg := <-outputCh:
			if !bytes.Equal(msg, []byte(expected)) {
				t.Errorf("message %d: expected '%s', got '%s'", i, expected, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("message %d: timeout waiting for echo", i)
		}
	}

	service.Stop()
}

func TestEchoServiceStop(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "echo:from-ws"
	writeTopic := "echo:to-ws"

	service := NewEchoService(bus, readTopic, writeTopic)
	service.Start(context.Background())

	outputCh := bus.Subscribe(writeTopic)

	bus.Publish(readTopic, []byte("before-stop"))

	select {
	case <-outputCh:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message before stop")
	}

	service.Stop()

	time.Sleep(50 * time.Millisecond)

	bus.Publish(readTopic, []byte("after-stop"))

	select {
	case msg := <-outputCh:
		t.Errorf("received message after stop: '%s'", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestEchoServiceContextCancellation(t *testing.T) {
	bus := messagebus.NewMessageBus()
	readTopic := "echo:from-ws"
	writeTopic := "echo:to-ws"

	service := NewEchoService(bus, readTopic, writeTopic)
	ctx, cancel := context.WithCancel(context.Background())

	service.Start(ctx)

	outputCh := bus.Subscribe(writeTopic)

	bus.Publish(readTopic, []byte("test"))

	select {
	case <-outputCh:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}

	cancel()

	time.Sleep(50 * time.Millisecond)

	bus.Publish(readTopic, []byte("after-cancel"))

	select {
	case msg := <-outputCh:
		t.Errorf("received message after context cancel: '%s'", msg)
	case <-time.After(100 * time.Millisecond):
	}
}
