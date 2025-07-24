package model

// QueryResult represents the response structure sent back to the client
type QueryResult struct {
	Columns []string                 `json:"columns"` // Column names from the query result
	Rows    []map[string]interface{} `json:"data"`    // Query result rows
}
