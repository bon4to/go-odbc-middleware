package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bon4to/go-odbc-middleware/internal/db"
	"github.com/bon4to/go-odbc-middleware/internal/model"
	"github.com/bon4to/go-odbc-middleware/internal/service"
)

// QueryHandler handles incoming HTTP POST requests to /query
func QueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data model.RequestData

	// Decode JSON body into RequestData struct
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Request from %s | Source: %d", r.RemoteAddr, data.Source)

	// Connect to the database using provided source index
	db, err := db.Connect(data.Source)
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Println("DB connection error:", err)
		return
	}
	defer db.Close()

	// Run the provided query
	result, err := service.Query(db, data.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
