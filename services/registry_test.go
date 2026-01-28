package services

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

type mockService struct {
	startCalled int32
	stopCalled  int32
	startErr    error
}

func (m *mockService) Start(ctx context.Context) error {
	atomic.AddInt32(&m.startCalled, 1)
	return m.startErr
}

func (m *mockService) Stop() error {
	atomic.AddInt32(&m.stopCalled, 1)
	return nil
}

func (m *mockService) started() int {
	return int(atomic.LoadInt32(&m.startCalled))
}

func (m *mockService) stopped() int {
	return int(atomic.LoadInt32(&m.stopCalled))
}

func TestRegistryAdd(t *testing.T) {
	bus := messagebus.NewInMemoryMessageBus()
	registry := NewServiceRegistry(bus)

	mock := &mockService{}
	err := registry.Add("test", mock)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mock.started() != 1 {
		t.Errorf("expected Start called once, got %d", mock.started())
	}

	if mock.stopped() != 0 {
		t.Errorf("expected Stop not called, got %d", mock.stopped())
	}
}

func TestRegistryGet(t *testing.T) {
	bus := messagebus.NewInMemoryMessageBus()
	registry := NewServiceRegistry(bus)

	mock := &mockService{}
	registry.Add("test", mock)

	svc, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if svc != mock {
		t.Error("expected same service instance")
	}

	if mock.started() != 1 {
		t.Errorf("expected Start called once, got %d", mock.started())
	}
}

func TestRegistryGetNonExistent(t *testing.T) {
	bus := messagebus.NewInMemoryMessageBus()
	registry := NewServiceRegistry(bus)

	svc, err := registry.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if svc != nil {
		t.Error("expected nil for nonexistent service")
	}
}

func TestRegistryReferenceCount(t *testing.T) {
	bus := messagebus.NewInMemoryMessageBus()
	registry := NewServiceRegistry(bus)

	mock := &mockService{}
	registry.Add("test", mock)

	registry.Get("test")
	registry.Get("test")

	if mock.stopped() != 0 {
		t.Error("service should not be stopped yet")
	}

	registry.Release("test")
	registry.Release("test")

	if mock.stopped() != 0 {
		t.Error("service should not be stopped yet")
	}

	registry.Release("test")

	time.Sleep(10 * time.Millisecond)

	if mock.stopped() != 1 {
		t.Errorf("expected Stop called once, got %d", mock.stopped())
	}
}

func TestRegistryReleaseNonExistent(t *testing.T) {
	bus := messagebus.NewInMemoryMessageBus()
	registry := NewServiceRegistry(bus)

	registry.Release("nonexistent")
}

func TestRegistryStopAll(t *testing.T) {
	bus := messagebus.NewInMemoryMessageBus()
	registry := NewServiceRegistry(bus)

	mock1 := &mockService{}
	mock2 := &mockService{}
	mock3 := &mockService{}

	registry.Add("svc1", mock1)
	registry.Add("svc2", mock2)
	registry.Add("svc3", mock3)

	registry.Get("svc1")
	registry.Get("svc2")

	registry.StopAll()

	time.Sleep(10 * time.Millisecond)

	if mock1.stopped() != 1 {
		t.Errorf("svc1: expected Stop called once, got %d", mock1.stopped())
	}
	if mock2.stopped() != 1 {
		t.Errorf("svc2: expected Stop called once, got %d", mock2.stopped())
	}
	if mock3.stopped() != 1 {
		t.Errorf("svc3: expected Stop called once, got %d", mock3.stopped())
	}

	svc, _ := registry.Get("svc1")
	if svc != nil {
		t.Error("expected all services cleared after StopAll")
	}
}
