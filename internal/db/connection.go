package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/bon4to/go-odbc-middleware/internal/config"
)

// Connect loads .env and connects to the database using source index
func Connect(sourceIndex int) (*sql.DB, error) {
	// Lookup database name based on source index
	dbSourceName := os.Getenv("DB_DSN_" + strconv.Itoa(sourceIndex))
	if dbSourceName == "" {
		return nil, fmt.Errorf("DB_DSN_%d not found in .env", sourceIndex)
	}

	config, err := config.Load(sourceIndex)
	if err != nil {
		return nil, err
	}

	// Format DSN string for go_ibm_db
	dsn := makeIBMString(config.Host, config.Port, dbSourceName, config.User, config.Password)

	// Open and validate the database connection
	db, err := openIBMConnection(dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
