package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/ibmdb/go_ibm_db"
	"github.com/joho/godotenv"
)

func main() {
	PrintTime("Building DSN...")
	dsn := dsnBuilder(getDotEnv())

	PrintTime("Connecting to Database...")
	db, err := sql.Open("go_ibm_db", dsn)
	if err != nil {
		log.Fatal("Unable to connect to Database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Error while testing connection with Database:", err)
	}
	PrintTime("Connection established.")

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		var requestBody struct {
			Query string `json:"query"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		}

		results, err := runQuery(db, requestBody.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error processing query: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	PrintTime("Starting on :40500")
	log.Fatal(http.ListenAndServe(":40500", nil))
}

// carrega os valores do .env
func getDotEnv() (string, string, string, string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to load the file .env.", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSourceName := os.Getenv("DB_DSN")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	return dbHost, dbPort, dbSourceName, dbUser, dbPassword
}

// imprime o log com timestamp
func PrintTime(text string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Print("[" + timestamp + "] " + text + "\n")
}

// monta a string de conexão com a dsn
func dsnBuilder(dbHost string, dbPort string, dbSourceName string, dbUser string, dbPassword string) string {
	dsn := "HOSTNAME=" + dbHost + ";PORT=" + dbPort + ";DATABASE=" + dbSourceName + ";UID=" + dbUser + ";PWD=" + dbPassword
	return dsn
}

func runQuery(db *sql.DB, query string) ([]map[string]interface{}, error) {
	start := time.Now()

	// processa query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// busca nome das colunas
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)

	for rows.Next() {
		// cria um slice de interfaces para armazenar os valores
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// lê a linha
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]

			switch b := val.(type) {
			case []byte:
				v = string(b)
			default:
				v = b
			}

			row[col] = v
		}
		results = append(results, row)
	}

	// verifica por erros na iteração
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// imprime o tempo de execução
	fmt.Printf("Query processed in %s\n", time.Since(start))

	return results, nil
}
