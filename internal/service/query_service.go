package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bon4to/go-odbc-middleware/internal/model"
)

// Query executes a SQL query and returns the result as QueryResult
func Query(db *sql.DB, query string) (*model.QueryResult, error) {
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

	log.Printf("Query completed in %s", time.Since(start))

	return &model.QueryResult{Columns: columns, Rows: results}, nil
}
