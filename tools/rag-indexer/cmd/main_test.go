package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectReadsOnlyApprovedTechnicalDocumentation(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "README.md"), "# Root\nroot document")
	mustWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\ndocumentation")
	mustWrite(t, filepath.Join(root, "ops", "runbooks", "health.md"), "# Health\nrunbook")
	mustWrite(t, filepath.Join(root, "contracts", "api.yaml"), "openapi: 3.1.0")
	mustWrite(t, filepath.Join(root, ".codex", "todo.md"), "private operational notes")
	mustWrite(t, filepath.Join(root, "env", "secret.env"), "SECRET=value")

	docs, err := collect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 4 {
		t.Fatalf("documents = %d, want 4: %#v", len(docs), docs)
	}
	for _, doc := range docs {
		if doc.Path == ".codex/todo.md" || doc.Path == "env/secret.env" {
			t.Fatalf("forbidden document indexed: %s", doc.Path)
		}
	}
}

func TestChunksForSplitsLongDocument(t *testing.T) {
	file := filepath.Join(t.TempDir(), "long.md")
	mustWrite(t, file, "# Long\n"+string(make([]byte, maxChunkBytes*2)))
	chunks := chunksFor(file, "docs/long.md")
	if len(chunks) < 2 {
		t.Fatalf("chunks = %d, want at least 2", len(chunks))
	}
	for _, chunk := range chunks {
		if len(chunk.Text) > maxChunkBytes {
			t.Fatalf("chunk length = %d", len(chunk.Text))
		}
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
