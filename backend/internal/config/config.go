// Package config loads runtime configuration from environment variables.
// Defaults are suitable for local development.
package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
}

// Load reads configuration from the environment, falling back to
// development defaults when a variable is not set.
func Load() Config {
	return Config{
		DatabaseURL: getEnv(
			"DATABASE_URL",
			"postgres://postgres:postgres@localhost:5432/schemaflow_dev?sslmode=disable",
		),
		Port: getEnv("PORT", "8080"),
	}
}

// getEnv returns the value of an environment variable, or a fallback
// if the variable is unset or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
