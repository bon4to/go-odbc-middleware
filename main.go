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

// RequestData represents the expected structure of the incoming JSON request body.
type RequestData struct {
	Query  string `json:"query"`
	Source int    `json:"source"`
}

func main() {
	http.HandleFunc("/query", handleQuery)

	log.Println("[INFO] Server is starting on port :40500...\n")
	log.Fatal(http.ListenAndServe(":40500", nil))
}

// handleQuery handles POST requests to the /query endpoint.
func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[WARN] Invalid method: %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data RequestData

	log.Printf("[INFO] Incoming request from %s", r.RemoteAddr)

	// Parse incoming JSON body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("[ERROR] Failed to decode JSON: %v", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Attempting to build DSN for source index %d", data.Source)
	dbHost, dbPort, dbSourceName, dbUser, dbPassword := getDotEnv(data.Source)
	dsn := dsnBuilder(dbHost, dbPort, dbSourceName, dbUser, dbPassword)

	// Establish database connection
	log.Print("[INFO] Connecting to the database...")
	db, err := sql.Open("go_ibm_db", dsn)
	if err != nil {
		log.Printf("[ERROR] Could not connect to DB (source %d): %v", data.Source, err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check database connection
	if err := db.Ping(); err != nil {
		log.Printf("[ERROR] DB ping failed (source %d): %v", data.Source, err)
		http.Error(w, "Database unreachable", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] Successfully connected to database (source %d)", data.Source)

	// Execute query and return results
	results, err := runQuery(db, data.Query)
	if err != nil {
		log.Printf("[ERROR] Query execution failed: %v", err)
		http.Error(w, fmt.Sprintf("Query execution error: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare and send JSON response
	response := struct {
		Columns []string                 `json:"columns"`
		Data    []map[string]interface{} `json:"data"`
	}{
		Columns: results.Columns,
		Data:    results.Rows,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode JSON response: %v", err)
	}
}

// getDotEnv loads the .env file and retrieves DB connection values for the given source index.
func getDotEnv(index int) (string, string, string, string, string) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("[FATAL] Failed to load .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dbSourceName, err := getDbSourceName(index)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	return dbHost, dbPort, dbSourceName, dbUser, dbPassword
}

// getDbSourceName retrieves the DSN string based on a given index (e.g., index=1 maps to DB_DSN_1).
func getDbSourceName(index int) (string, error) {
	dbSourceName := os.Getenv("DB_DSN_" + strconv.Itoa(index))
	if dbSourceName == "" {
		return "", fmt.Errorf("DB source name for index %d not found", index)
	}
	return dbSourceName, nil
}

// dsnBuilder builds a DSN string for IBM DB2 based on environment values.
func dsnBuilder(dbHost, dbPort, dbSourceName, dbUser, dbPassword string) string {
	return fmt.Sprintf(
		"HOSTNAME=%s;PORT=%s;DATABASE=%s;UID=%s;PWD=%s",
		dbHost, dbPort, dbSourceName, dbUser, dbPassword,
	)
}

// QueryResult holds the columns and row data returned from a SQL query.
type QueryResult struct {
	Columns []string
	Rows    []map[string]interface{}
}

// runQuery executes the SQL query and returns results as a slice of maps.
func runQuery(db *sql.DB, queryString string) (*QueryResult, error) {
	if queryString == "" {
		return nil, fmt.Errorf("query string is empty")
	}

	start := time.Now()
	rows, err := db.Query(queryString)
	if err != nil {
		log.Printf("[ERROR] SQL execution failed: %v", err)
		return nil, fmt.Errorf("query execution failed")
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve column names: %v", err)
		return nil, err
	}

	results := make([]map[string]interface{}, 0)

	// Iterate over each row and scan values into a map
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("[ERROR] Failed to scan row: %v", err)
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Row iteration error: %v", err)
		return nil, err
	}

	log.Printf("[INFO] Query executed in %s\n", time.Since(start))
	return &QueryResult{Columns: columns, Rows: results}, nil
}
