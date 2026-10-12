// database.go
// Package main contains the database connection logic for SchemaFlow
package main

import (
	"context"                         // Controls running operations (like timeouts, cancellations)
	"github.com/jackc/pgx/v5/pgxpool" // PostgreSQL connection pool driver
	"log"                             // Prints messages to terminal
)

// DB is a global variable that holds the database connection pool
// Any function can use this to query the database
var DB *pgxpool.Pool

// InitDB creates a connection to PostgreSQL and stores it in the DB variable
// It must be called before using the database (usually in main())
func InitDB() error {
	// STEP 1: Define connection string
	// Format: postgres://username:password@host:port/database?sslmode=disable
	connString := "postgres://postgres:postgres@localhost:5432/schemaflow_dev?sslmode=disable"

	// STEP 2: Connect to PostgreSQL using pgxpool (connection pool)
	var err error
	DB, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return err
	}

	// STEP 3: Verify the connection actually works by pinging the database
	err = DB.Ping(context.Background())
	if err != nil {
		return err
	}

	// STEP 4: Log success
	log.Println("Database connection pool created successfully")

	// STEP 5: Return nil (no error) meaning success
	return nil
}
