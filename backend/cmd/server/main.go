package main

import (
	"context"
	"log"
	"net/http"

	"schemaflow-backend/internal/config"
	"schemaflow-backend/internal/db"
	"schemaflow-backend/internal/handler"
	"schemaflow-backend/internal/httpx"
	"schemaflow-backend/internal/repository"
	"schemaflow-backend/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()
	log.Println("Database connection pool created successfully")

	// repositories
	eventsRepo := repository.NewSchemaEventsRepo(pool)
	lineageRepo := repository.NewColumnLineageRepo(pool)

	// services
	migrateService := service.NewMigrateService(eventsRepo, lineageRepo)
	rewriterService := service.NewRewriterService(pool)
	impactService := service.NewImpactService(pool)

	h := handler.New(migrateService, rewriterService, impactService, eventsRepo, lineageRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/migrate", h.MigrateHandler)
	mux.HandleFunc("/parse", h.ParseHandler)
	mux.HandleFunc("/rewrite", h.RewriteHandler)
	mux.HandleFunc("/impact", h.ImpactHandler)
	mux.HandleFunc("/migrations", h.MigrationsHandler)
	mux.HandleFunc("/lineage", h.LineageHandler)

	stack := httpx.Logging(httpx.CORS(mux))

	log.Println("Server starting on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, stack))
}
