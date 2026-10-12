// main.go - This file starts the HTTP server and defines all API endpoints
// package main is the entry point for the SchemaFlow backend server
package main

import (
	"encoding/json" // Converts Go structs to JSON and JSON back to Go structs
	"fmt"           // Prints messages to the terminal using Println and Printf
	"log"           // Prints error and info messages with date and time to the terminal
	"net/http"      // Creates an HTTP server and handles incoming client requests
	"strings"       // Converts strings to uppercase using ToUpper and other text operations
)

// main is the first function that runs when the program starts
// It connects to the database, sets up API routes, and starts the server
func main() {
	// STEP 1: Connect to PostgreSQL database
	err := InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to PostgreSQL")

	// STEP 2: Define API routes (endpoints)
	// Each route handles a specific functionality of SchemaFlow
	http.HandleFunc("/migrate", corsMiddleware(MigrateHandler))
	http.HandleFunc("/parse", corsMiddleware(ParseHandler))
	http.HandleFunc("/rewrite", corsMiddleware(RewriteHandler))
	http.HandleFunc("/impact", corsMiddleware(ImpactHandler))
	http.HandleFunc("/migrations", corsMiddleware(MigrationsHandler))
	http.HandleFunc("/lineage", corsMiddleware(LineageHandler))

	// STEP 3: Start the HTTP server on port 8080
	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HANDLER 1: /migrate - Record a migration event
// MigrateHandler handles POST requests to /migrate
// It receives migration details and saves them to the database
// Example: {"event_type":"rename_column","table_name":"users","old_name":"full_name","new_name":"buyer_name","applied_by":"deepshan"}
func MigrateHandler(w http.ResponseWriter, r *http.Request) {
	// STEP 1: Allow only POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST allowed"))
		return
	}

	// STEP 2: Define the expected JSON structure
	var req struct {
		EventType string `json:"event_type"` // e.g., "rename_column", "add_column", "drop_column"
		TableName string `json:"table_name"` // e.g., "users"
		OldName   string `json:"old_name"`   // Old column name (for rename)
		NewName   string `json:"new_name"`   // New column name (for rename)
		AppliedBy string `json:"applied_by"` // Who applied the migration
	}

	// STEP 3: Parse the JSON from the request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	// STEP 4: Normalize event_type to uppercase
	// PostgreSQL check constraint expects RENAME_COLUMN, ADD_COLUMN, or DROP_COLUMN
	req.EventType = strings.ToUpper(req.EventType)

	// STEP 5: Validate required fields
	if req.EventType == "" || req.TableName == "" || req.AppliedBy == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required fields: event_type, table_name, applied_by"))
		return
	}

	// STEP 6: Save to database (calls the function in migration.go)
	err = SaveMigrationEvent(req.EventType, req.TableName, req.OldName, req.NewName, req.AppliedBy)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to save migration: " + err.Error()))
		return
	}

	// STEP 7: Send success response
	resp := map[string]string{
		"status":  "saved",
		"message": "Migration recorded successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HANDLER 2: /parse - Parse a SQL query and extract columns and tables
// ParseHandler handles POST requests to /parse
// It parses a SQL query and returns the extracted column and table names
// Example: {"query":"SELECT full_name, email FROM users WHERE id = 123"}
func ParseHandler(w http.ResponseWriter, r *http.Request) {
	// STEP 1: Allow only POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST allowed"))
		return
	}

	// STEP 2: Define the expected JSON structure
	var req struct {
		Query string `json:"query"` // The SQL query to parse
	}

	// STEP 3: Parse the JSON from the request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	// STEP 4: Validate required fields
	if req.Query == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required field: query"))
		return
	}

	// STEP 5: Parse the query using the rewriter package
	parsed, err := ParseQuery(req.Query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse query: " + err.Error()))
		return
	}

	// STEP 6: Send response with extracted columns and tables
	resp := map[string]interface{}{
		"columns": parsed.GetColumnNames(), // List of column names found
		"tables":  parsed.GetTableNames(),  // List of table names found
		"counts": map[string]int{
			"columns": len(parsed.Columns), // Number of columns found
			"tables":  len(parsed.Tables),  // Number of tables found
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HANDLER 3: /rewrite - Rewrite a SQL query with new column names
// RewriteHandler handles POST requests to /rewrite
// It rewrites a SQL query by replacing old column names with new ones from column_lineage
// Example: {"query":"SELECT full_name, email FROM users WHERE id = 123"}
// Response: {"original":"SELECT full_name, email FROM users WHERE id = 123","rewritten":"SELECT buyer_name, email FROM users WHERE id = 123"}
func RewriteHandler(w http.ResponseWriter, r *http.Request) {
	// STEP 1: Allow only POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST allowed"))
		return
	}

	// STEP 2: Define the expected JSON structure
	var req struct {
		Query string `json:"query"` // The SQL query to rewrite
	}

	// STEP 3: Parse the JSON from the request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	// STEP 4: Validate required fields
	if req.Query == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required field: query"))
		return
	}

	// STEP 5: Rewrite the query using the rewriter package
	rewritten, err := RewriteQuery(req.Query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to rewrite query: " + err.Error()))
		return
	}

	// STEP 6: Send response with original and rewritten query
	resp := map[string]interface{}{
		"original":  req.Query, // The original query sent by the application
		"rewritten": rewritten, // The rewritten query with new column names
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HANDLER 4: /impact - Predict impact of a migration (Step 5)
// ImpactHandler handles POST requests to /impact
// It queries PostgreSQL's system catalog to find objects that depend on a column
// Example: {"schema":"public","table":"users","column":"full_name"}
func ImpactHandler(w http.ResponseWriter, r *http.Request) {
	// STEP 1: Allow only POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST allowed"))
		return
	}

	// STEP 2: Define the expected JSON structure
	var req struct {
		Schema string `json:"schema"`
		Table  string `json:"table"`
		Column string `json:"column"`
	}

	// STEP 3: Parse the JSON from the request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	// STEP 4: Validate required fields
	if req.Schema == "" {
		req.Schema = "public" // Default schema
	}
	if req.Table == "" || req.Column == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required fields: table, column"))
		return
	}

	// STEP 5: Get dependencies from the database
	report, err := GetColumnDependencies(req.Schema, req.Table, req.Column)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get dependencies: " + err.Error()))
		return
	}

	// STEP 6: Build the response
	resp := map[string]interface{}{
		"column":           req.Column,
		"table":            req.Table,
		"schema":           req.Schema,
		"dependencies":     report.Dependencies,
		"total":            report.Total,
		"has_dependencies": report.HasDependencies(),
	}

	// Add warning if dependencies exist
	if report.HasDependencies() {
		resp["warning"] = fmt.Sprintf("%d dependencies found. Renaming this column may break them.", report.Total)
	} else {
		resp["message"] = "No dependencies found. Safe to rename this column."
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// MigrationsHandler returns all migrations from schema_events.
func MigrationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only GET allowed"))
		return
	}

	migrations, err := ListMigrations()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to list migrations: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(migrations)
}

// LineageHandler returns all column lineage records.
func LineageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only GET allowed"))
		return
	}

	records, err := ListLineage()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to list lineage: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// corsMiddleware adds CORS headers and handles preflight OPTIONS requests.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request — browser sends this before POST
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
