package main

import "testing"

func TestGenreLabel(t *testing.T) {
	if got := genreLabel("sci-fi"); got != "Sci-fi" {
		t.Fatalf("genreLabel = %q", got)
	}
}

func TestNullIfEmpty(t *testing.T) {
	if got := nullIfEmpty("   "); got != nil {
		t.Fatalf("nullIfEmpty blank = %#v", got)
	}

	if got := nullIfEmpty("Arrival"); got != "Arrival" {
		t.Fatalf("nullIfEmpty value = %#v", got)
	}
}
