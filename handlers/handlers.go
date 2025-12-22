package handlers

import (
	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
	"github.com/samuel1992/ws-server-with-messagebus/services"
	"net/http"

	"github.com/gorilla/websocket"
)

type Handlers struct {
	registry *services.ServiceRegistry
	bus      messagebus.MessageBus
	upgrader websocket.Upgrader
}

func NewHandlers(registry *services.ServiceRegistry, bus messagebus.MessageBus) *Handlers {
	return &Handlers{
		registry: registry,
		bus:      bus,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}
