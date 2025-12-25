package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
	"github.com/samuel1992/ws-server-with-messagebus/services"
	"github.com/samuel1992/ws-server-with-messagebus/ws"

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

type ServiceFactory func(bus messagebus.MessageBus, fromWsToService, fromServiceToWs string) services.Service

func (h *Handlers) HandleWs(w http.ResponseWriter, r *http.Request, serviceFactory ServiceFactory) {
	endpoint := strings.TrimPrefix(r.URL.Path, "/ws/")

	fromWsToService := endpoint + ":from-ws-to-service"
	fromServiceToWs := endpoint + ":from-service-to-ws"

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	service, err := h.registry.Get(endpoint)
	if err != nil {
		log.Println("Error getting service:", err)
		conn.Close()
		return
	}
	if service == nil {
		newService := serviceFactory(h.bus, fromWsToService, fromServiceToWs)
		h.registry.Add(endpoint, newService)
	}

	wsClient := ws.NewClient(conn, h.bus, fromServiceToWs, fromWsToService)

	defer func() {
		log.Println("Cleaning up service resources")
		wsClient.Stop()
		h.registry.Release(endpoint)
	}()

	err = wsClient.Start()
	if err != nil {
		log.Println("Error starting WS client:", err)
		wsClient.Stop()
		h.registry.Release(endpoint)
		return
	}
}
