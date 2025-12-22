package main

import (
	"log"
	"net/http"

	"github.com/samuel1992/ws-server-with-messagebus/messagebus"
	"github.com/samuel1992/ws-server-with-messagebus/services"

	"github.com/samuel1992/ws-server-with-messagebus/handlers"
)

func main() {
	port := ":3000"

	messageBus := messagebus.NewMessageBus()
	serviceRegistry := services.NewServiceRegistry(messageBus)
	handlers := handlers.NewHandlers(serviceRegistry, messageBus)

	// The handler now accepts the message bus and service registry.
	http.HandleFunc("/ws/ping", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlePing(w, r)
	})

	http.HandleFunc("/ws/timenow", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleTimeNow(w, r)
	})

	log.Printf("Starting server on %v\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
