package main

import "testing"

func TestGetenv(t *testing.T) {
	t.Setenv("MOVIE_CATALOG_MIGRATION_FILE", "custom.sql")

	if got := getenv("MOVIE_CATALOG_MIGRATION_FILE", "fallback.sql"); got != "custom.sql" {
		t.Fatalf("getenv = %q", got)
	}
}
