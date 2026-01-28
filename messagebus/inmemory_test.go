package messagebus

import (
	"bytes"
	"testing"
	"time"
)

func TestBasicPubSub(t *testing.T) {
	bus := NewInMemoryMessageBus()
	topic := "test-topic"

	ch := bus.Subscribe(topic)
	bus.Publish(topic, []byte("hello"))

	select {
	case msg := <-ch:
		if !bytes.Equal(msg, []byte("hello")) {
			t.Errorf("expected 'hello', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}
}

func TestMultipleSubscribersSameTopic(t *testing.T) {
	bus := NewInMemoryMessageBus()
	topic := "test-topic"

	ch1 := bus.Subscribe(topic)
	ch2 := bus.Subscribe(topic)
	ch3 := bus.Subscribe(topic)

	bus.Publish(topic, []byte("broadcast"))

	channels := []chan []byte{ch1, ch2, ch3}
	for i, ch := range channels {
		select {
		case msg := <-ch:
			if !bytes.Equal(msg, []byte("broadcast")) {
				t.Errorf("subscriber %d: expected 'broadcast', got '%s'", i, msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("subscriber %d: timeout waiting for message", i)
		}
	}
}

func TestTopicIsolation(t *testing.T) {
	bus := NewInMemoryMessageBus()

	ch1 := bus.Subscribe("topic1")
	ch2 := bus.Subscribe("topic2")

	bus.Publish("topic1", []byte("message1"))
	bus.Publish("topic2", []byte("message2"))

	select {
	case msg := <-ch1:
		if !bytes.Equal(msg, []byte("message1")) {
			t.Errorf("topic1: expected 'message1', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("topic1: timeout waiting for message")
	}

	select {
	case msg := <-ch2:
		if !bytes.Equal(msg, []byte("message2")) {
			t.Errorf("topic2: expected 'message2', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("topic2: timeout waiting for message")
	}

	select {
	case msg := <-ch1:
		t.Errorf("topic1 received unexpected message: '%s'", msg)
	case <-time.After(50 * time.Millisecond):
	}

	select {
	case msg := <-ch2:
		t.Errorf("topic2 received unexpected message: '%s'", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewInMemoryMessageBus()
	topic := "test-topic"

	ch := bus.Subscribe(topic)
	bus.Publish(topic, []byte("before"))

	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message before unsubscribe")
	}

	bus.Unsubscribe(topic, ch)

	_, ok := <-ch
	if ok {
		t.Fatal("channel should be closed after unsubscribe")
	}

	bus.Publish(topic, []byte("after"))

	select {
	case msg, ok := <-ch:
		if ok {
			t.Errorf("received message after unsubscribe: '%s'", msg)
		}
	case <-time.After(50 * time.Millisecond):
	}
}

func TestFullChannelDropsMessage(t *testing.T) {
	bus := NewInMemoryMessageBus()
	topic := "test-topic"

	ch := bus.Subscribe(topic)

	for i := 0; i < 256; i++ {
		bus.Publish(topic, []byte("fill"))
	}

	bus.Publish(topic, []byte("dropped"))

	time.Sleep(10 * time.Millisecond)

	received := 0
	for len(ch) > 0 {
		<-ch
		received++
	}

	if received != 256 {
		t.Errorf("expected 256 messages in channel, got %d", received)
	}
}

func TestSubscribeAfterPublish(t *testing.T) {
	bus := NewInMemoryMessageBus()
	topic := "test-topic"

	bus.Publish(topic, []byte("old"))

	ch := bus.Subscribe(topic)

	select {
	case msg := <-ch:
		t.Errorf("received old message: '%s'", msg)
	case <-time.After(50 * time.Millisecond):
	}

	bus.Publish(topic, []byte("new"))

	select {
	case msg := <-ch:
		if !bytes.Equal(msg, []byte("new")) {
			t.Errorf("expected 'new', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for new message")
	}
}

func TestPublishToTopicWithNoSubscribers(t *testing.T) {
	bus := NewInMemoryMessageBus()

	bus.Publish("nonexistent", []byte("test"))
}
