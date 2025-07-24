package main

import (
	"log"
	"net/http"

	"github.com/bon4to/go-odbc-middleware/internal/handler"
)

func main() {
	// Handle POST requests at /query endpoint
	http.HandleFunc("/query", handler.QueryHandler)

	port := ":40500"
	log.Printf("Server running on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
