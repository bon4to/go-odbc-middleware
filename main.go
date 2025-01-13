package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/ibmdb/go_ibm_db"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("o arquivo .env não pôde ser carregado.", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dsn := dsn_builder(dbHost, dbPort, dbUser, dbPassword)
	fmt.Println("dsn configurada:", dsn)
}

func dsn_builder(dbHost string, dbPort string, dbUser string, dbPassword string) string {
	dsn := "HOSTNAME=" + dbHost + ";PORT=" + dbPort + ";DATABASE=HUGOPIET;UID=" + dbUser + ";PWD=" + dbPassword
	return dsn
}
