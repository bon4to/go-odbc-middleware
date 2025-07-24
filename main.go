package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/ibmdb/go_ibm_db"
	"github.com/joho/godotenv"
)

// RequestData represents the expected JSON structure from the client
type RequestData struct {
	Query  string `json:"query"`  // SQL query to execute
	Source int    `json:"source"` // DB source index (for selecting credentials)
}

// QueryResult represents the response structure sent back to the client
type QueryResult struct {
	Columns []string                 `json:"columns"` // Column names from the query result
	Rows    []map[string]interface{} `json:"data"`    // Query result rows
}

func main() {
	// Handle POST requests at /query endpoint
	http.HandleFunc("/query", queryHandler)

	port := ":40500"
	logRequest("CDE", "Server running on: "+port, http.StatusOK, nil, 0)
	log.Fatal(http.ListenAndServe(port, nil))
}

func logRequest(level string, text string, status int, r *http.Request, source int) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if r == nil {
		fmt.Printf("[%s] (%s) %s - %d\n", level, timestamp, text, status)
		return
	}
	ip := r.RemoteAddr
	fmt.Printf("[%s] (%s) %s - %d | %s (%d - %s)\n", level, timestamp, text, status, ip, source, r.URL.Path)
	return
}

// queryHandler handles incoming HTTP POST requests to /query
func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data RequestData

	// Decode JSON body into RequestData struct
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	logRequest("INF", "Query received", http.StatusOK, r, data.Source)

	// Connect to the database using provided source index
	db, err := connectDB(data.Source)
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		logRequest("ERR", "Database connection error:", http.StatusInternalServerError, r, data.Source)
		log.Println("\n\n" + err.Error() + "\n\n") // log does a stderr write according to https://pkg.go.dev/log
		return
	}
	defer db.Close()

	// Run the provided query
	result, err := runQuery(db, data.Query)
	if err != nil {
		logRequest("ERR", "Query execution error:", http.StatusInternalServerError, r, data.Source)
		log.Println("\n\n" + err.Error() + "\n\n") // stderr write
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// connectDB loads .env and connects to the database using source index
func connectDB(sourceIndex int) (*sql.DB, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	// Read general connection settings
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// Lookup database name based on source index
	dbSourceName := os.Getenv("DB_DSN_" + strconv.Itoa(sourceIndex))
	if dbSourceName == "" {
		return nil, fmt.Errorf("DB_DSN_%d not found in .env", sourceIndex)
	}

	// Format DSN string for go_ibm_db
	dsn := fmt.Sprintf(
		"HOSTNAME=%s;PORT=%s;DATABASE=%s;UID=%s;PWD=%s",
		dbHost, dbPort, dbSourceName, dbUser, dbPassword,
	)

	// Open and validate the database connection
	db, err := sql.Open("go_ibm_db", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// runQuery executes a SQL query and returns the result as QueryResult
func runQuery(db *sql.DB, query string) (*QueryResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	start := time.Now()
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)

	// Iterate over each row and build result set
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert byte slices to strings for JSON encoding
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	// Return any scanning errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	logRequest("INF", "Query completed in: "+time.Since(start).String(), http.StatusOK, nil, 0)

	return &QueryResult{Columns: columns, Rows: results}, nil
}
