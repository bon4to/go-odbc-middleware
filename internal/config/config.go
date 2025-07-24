package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("failed to load .env: ", err)
	}
}

// Config represents the database connection settings
type Config struct {
	Host     string
	Port     string
	Source   string
	User     string
	Password string
}

// New creates a new Config struct
func New(host, port, source, user, password string) Config {
	return Config{
		Host:     host,
		Port:     port,
		Source:   source,
		User:     user,
		Password: password,
	}
}

// Load loads the database connection settings from environment variables
func Load(sourceIndex int) (Config, error) {
	// Read general connection settings
	Host := os.Getenv("DB_HOST")
	Port := os.Getenv("DB_PORT")
	User := os.Getenv("DB_USER")
	Password := os.Getenv("DB_PASSWORD")

	// Lookup .env for dsn by source index
	SourceName := os.Getenv("DB_DSN_" + strconv.Itoa(sourceIndex))
	if SourceName == "" {
		return Config{}, fmt.Errorf("DB_DSN_%d not found in .env", sourceIndex)
	}

	return New(Host, Port, SourceName, User, Password), nil
}
