package model

// RequestData represents the expected JSON structure from the client
type RequestData struct {
	Query  string `json:"query"`  // SQL query to execute
	Source int    `json:"source"` // DB source index (for selecting credentials)
}
