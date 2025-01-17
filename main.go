package main

import (
	"database/sql"
	"fmt"
	"log"
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
}

// carrega os valores do .env
func getDotEnv() (string, string, string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to load the file .env.", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	return dbHost, dbPort, dbUser, dbPassword
}

// imprime o log com timestamp
func PrintTime(text string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Print("[" + timestamp + "] " + text + "\n")
}

// monta a string de conexão com a dsn
func dsnBuilder(dbHost string, dbPort string, dbUser string, dbPassword string) string {
	dsn := "HOSTNAME=" + dbHost + ";PORT=" + dbPort + ";DATABASE=HUGOPIET;UID=" + dbUser + ";PWD=" + dbPassword
	return dsn
}

func runQuery(db *sql.DB, query string) {
	start := time.Now()

	// processa query
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error while processing query:", err)
	}
	defer rows.Close()

	// busca nome das colunas
	columns, err := rows.Columns()
	if err != nil {
		log.Fatal("Error while getting column names:", err)
	}

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
			log.Fatal("Error while scanning results: ", err)
		}

		// imprime os valores
		// result := make(map[string]interface{})
		// for i, col := range columns {
		// 	val := values[i]
		// 	result[col] = val
		// }
		// fmt.Println(result)
	}
	// verifica por erros na iteração
	if err = rows.Err(); err != nil {
		log.Fatal("Error while iterating: ", err)
	}

	// imprime o tempo de execução
	fmt.Printf("Query processed in %s\n", time.Since(start))
}
