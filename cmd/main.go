package main

import (
	"fmt"
	"log"
	"net/http"

	"hot-cofee/internal/config"
	"hot-cofee/internal/handler"
)

func init() {
	if err := config.ConfigLoad(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	port := config.GetConfigPort()
	mux := http.NewServeMux()

	handler.InventoryEndpoints(mux)
	handler.MenuEndpoints(mux)
	handler.OrderEndpoints(mux)
	handler.AggregationEndpoints(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.ErrorResponse(w, "405 - No such method", http.StatusMethodNotAllowed)
	})

	fmt.Println("Server started listening on port -", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
