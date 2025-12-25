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
	handler := handlers.NewWSHandler(serviceRegistry, messageBus)

	http.HandleFunc("/ws/ping", func(w http.ResponseWriter, r *http.Request) {
		handler.Handle(w, r, services.NewPingService)
	})

	http.HandleFunc("/ws/timenow", func(w http.ResponseWriter, r *http.Request) {
		handler.Handle(w, r, services.NewTimeNowService)
	})

	log.Printf("Starting server on %v\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
