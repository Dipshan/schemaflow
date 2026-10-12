// concurrency_test.go
// Experiment 4: Concurrency Test
// Runs 10 goroutines reading from title_basics while a migration happens

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// STEP 1: Connect to PostgreSQL
	connString := "postgres://postgres:postgres@localhost:5432/schemaflow_dev?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer pool.Close()

	fmt.Println("Concurrency Experiment Started...")
	fmt.Println("Starting 10 goroutines reading from title_basics...")

	// STEP 2: Track results
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	readsCompleted := 0
	var mu sync.Mutex

	startTime := time.Now()

	// STEP 3: Start 10 goroutines
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 1; j <= 100; j++ {
				var tconst, movieTitle string
				err := pool.QueryRow(context.Background(),
					"SELECT tconst, movie_title FROM title_basics ORDER BY random() LIMIT 1",
				).Scan(&tconst, &movieTitle)

				if err != nil {
					errors <- fmt.Errorf("goroutine %d read %d failed: %v", goroutineID, j, err)
					return
				}

				mu.Lock()
				readsCompleted++
				mu.Unlock()
			}

			fmt.Printf("Goroutine %d: 100 reads complete\n", goroutineID)
		}(i)
	}

	// STEP 4: Run migration in the middle
	time.Sleep(500 * time.Millisecond)
	fmt.Println(">>> Running migration: ALTER TABLE ADD COLUMN <<<")

	_, err = pool.Exec(context.Background(),
		"ALTER TABLE title_basics ADD COLUMN IF NOT EXISTS concurrency_test_col INTEGER")
	if err != nil {
		log.Printf("Migration error: %v", err)
	} else {
		fmt.Println(">>> Migration complete <<<")
	}

	// STEP 5: Wait for all goroutines
	wg.Wait()
	close(errors)

	// STEP 6: Report
	elapsed := time.Since(startTime)
	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("Total reads completed: %d / 1000\n", readsCompleted)
	fmt.Printf("Total errors: %d\n", len(errors))
	fmt.Printf("Elapsed time: %v\n", elapsed)

	if len(errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for err := range errors {
			fmt.Println(" -", err)
		}
		fmt.Println("\nFAIL: Errors occurred during migration")
	} else if readsCompleted == 1000 {
		fmt.Println("\nPASS: No deadlocks, no data loss")
	} else {
		fmt.Println("\nFAIL: Some reads did not complete")
	}

	// STEP 7: Cleanup (optional)
	fmt.Println("\nCleaning up test column...")
	_, err = pool.Exec(context.Background(),
		"ALTER TABLE title_basics DROP COLUMN IF EXISTS concurrency_test_col")
	if err != nil {
		log.Printf("Cleanup error: %v", err)
	}
	fmt.Println("Done.")
}
