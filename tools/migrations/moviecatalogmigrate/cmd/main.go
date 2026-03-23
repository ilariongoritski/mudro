package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDSN = "postgres://postgres:postgres@localhost:5434/movie_catalog?sslmode=disable"
const defaultMigrationFile = "migrations/movie_catalog/0001_init.sql"

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	dsn := getenv("MOVIE_CATALOG_DB_DSN", defaultDSN)
	migrationFile := getenv("MOVIE_CATALOG_MIGRATION_FILE", defaultMigrationFile)

	sqlBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	fmt.Println("movie catalog migration applied")
	return nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
