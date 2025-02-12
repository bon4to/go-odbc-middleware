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

type RequestData struct {
	Query  string `json:"query"`
	Source int    `json:"source"`
}

func main() {
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var data RequestData

		remoteAddr := r.RemoteAddr
		log.Printf("Received request from %s", remoteAddr)

		// Decodificar o JSON recebido
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
		}

		// Construir DSN usando o source enviado
		log.Printf("Building DSN for source %d...", data.Source)
		dbHost, dbPort, dbSourceName, dbUser, dbPassword := getDotEnv(data.Source)
		dsn := dsnBuilder(dbHost, dbPort, dbSourceName, dbUser, dbPassword)

		// Conectar ao banco de dados
		log.Print("Connecting to Database...")
		db, err := sql.Open("go_ibm_db", dsn)
		if err != nil {
			log.Printf("Unable to connect to Database for source %d: %v", data.Source, err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Printf("Database ping error for source %d: %v", data.Source, err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		log.Printf("Connection established for source %d.", data.Source)

		// Executar a query
		results, err := runQuery(db, data.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error processing query: %v", err), http.StatusInternalServerError)
			return
		}

		// Enviar a resposta
		response := struct {
			Columns []string                 `json:"columns"`
			Data    []map[string]interface{} `json:"data"`
		}{
			Columns: results.Columns,
			Data:    results.Rows,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("Starting on :40500")
	log.Fatal(http.ListenAndServe(":40500", nil))
}

/*
função para carregar os valores do .env

	recebe o índice correspondente ao nome da DSN no .env
	é utilizado em getDbSourceName(index)
*/
func getDotEnv(index int) (string, string, string, string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to load the file .env.", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dbSourceName, err := getDbSourceName(index)
	if err != nil {
		log.Fatal(err)
	}

	return dbHost, dbPort, dbSourceName, dbUser, dbPassword
}

/*
monta a string da dsn

	recebe o índice correspondente ao nome da DSN no .env
	exemplo: index = 1 --> 'DB_DSN_1'
*/
func getDbSourceName(index int) (string, error) {
	dbSourceName := os.Getenv("DB_DSN_" + strconv.Itoa(index))
	if dbSourceName == "" {
		return "", fmt.Errorf("invalid DB source name for index %d", index)
	}
	return dbSourceName, nil
}

// monta a string de conexão com a dsn
func dsnBuilder(dbHost string, dbPort string, dbSourceName string, dbUser string, dbPassword string) string {
	dsn := "HOSTNAME=" + dbHost + ";PORT=" + dbPort + ";DATABASE=" + dbSourceName + ";UID=" + dbUser + ";PWD=" + dbPassword
	return dsn
}

type QueryResult struct {
	Columns []string
	Rows    []map[string]interface{}
}

func runQuery(db *sql.DB, queryString string) (*QueryResult, error) {
	if queryString == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	start := time.Now()
	rows, err := db.Query(queryString)
	if err != nil {
		log.Printf("Query execution error: %v", err)
		return nil, fmt.Errorf("failed to execute query")
	}
	defer rows.Close()

	// obtem as colunas
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// imprime o tempo de execução
	fmt.Printf("Query processed in %s\n", time.Since(start))

	return &QueryResult{Columns: columns, Rows: results}, nil
}
