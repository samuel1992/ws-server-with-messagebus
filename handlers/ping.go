package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/samuel1992/ws-server-with-messagebus/services"
	"github.com/samuel1992/ws-server-with-messagebus/ws"
)

func (h *Handlers) HandlePing(w http.ResponseWriter, r *http.Request) {
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
		newService := services.NewPingService(h.bus, fromWsToService, fromServiceToWs)
		h.registry.Add(endpoint, newService)
	}

	wsClient := ws.NewClient(conn, h.bus, fromServiceToWs, fromWsToService)

	defer func() {
		log.Println("Cleaning up HandlePing resources")
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
