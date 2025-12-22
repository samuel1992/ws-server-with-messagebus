package services

import (
	"context"
	"log"
	"sync"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
)

type Service interface {
	Start(ctx context.Context) error
	Stop() error
}

type ServiceEntry struct {
	service  Service
	refCount int32
	mu       sync.Mutex
}

type ServiceRegistry struct {
	bus      messagebus.MessageBus
	services map[string]*ServiceEntry
	mu       sync.RWMutex
}

func NewServiceRegistry(bus messagebus.MessageBus) *ServiceRegistry {
	return &ServiceRegistry{
		bus:      bus,
		services: make(map[string]*ServiceEntry),
	}
}

func (r *ServiceRegistry) Add(endpoint string, service Service) error {
	// Start the provided service
	bgCtx := context.Background()
	err := service.Start(bgCtx)
	if err != nil {
		return err
	}

	entry := &ServiceEntry{
		service:  service,
		refCount: 1,
	}

	r.mu.Lock()
	r.services[endpoint] = entry
	r.mu.Unlock()

	log.Printf("Service %s: created (refCount now 1)\n", endpoint)
	return nil
}

// GetOrAdd returns an existing service or adds the provided one
func (r *ServiceRegistry) Get(endpoint string) (Service, error) {
	r.mu.RLock()
	entry, exists := r.services[endpoint]
	r.mu.RUnlock()

	if exists {
		entry.mu.Lock()
		entry.refCount++
		refCount := entry.refCount
		entry.mu.Unlock()
		log.Printf("Service %s: acquired (refCount now %d)\n", endpoint, refCount)

		return entry.service, nil
	}

	return nil, nil
}

func (r *ServiceRegistry) Release(endpoint string) {
	r.mu.RLock()
	entry, exists := r.services[endpoint]
	r.mu.RUnlock()

	if !exists {
		log.Printf("Service %s: release called but not found\n", endpoint)
		return
	}

	entry.mu.Lock()
	entry.refCount--
	refCount := entry.refCount
	entry.mu.Unlock()

	if refCount <= 0 {
		// Last client disconnected
		r.mu.Lock()
		delete(r.services, endpoint)
		r.mu.Unlock()

		entry.service.Stop()
		log.Printf("Service %s: stopped (refCount was %d)\n", endpoint, refCount)
	} else {
		log.Printf("Service %s: released (refCount now %d)\n", endpoint, refCount)
	}
}

// StopAll stops all services regardless of reference count
func (r *ServiceRegistry) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for endpoint, entry := range r.services {
		entry.service.Stop()
		log.Printf("Service %s: stopped during shutdown\n", endpoint)
	}

	r.services = make(map[string]*ServiceEntry)
}
