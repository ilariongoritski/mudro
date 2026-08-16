package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollectReadsOnlyApprovedTechnicalDocumentation(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "README.md"), "# Root\nroot document")
	mustWrite(t, filepath.Join(root, "docs", "guide.md"), "# Guide\ndocumentation")
	mustWrite(t, filepath.Join(root, "ops", "runbooks", "health.md"), "# Health\nrunbook")
	mustWrite(t, filepath.Join(root, "contracts", "api.yaml"), "openapi: 3.1.0")
	mustWrite(t, filepath.Join(root, ".codex", "todo.md"), "private operational notes")
	mustWrite(t, filepath.Join(root, "env", "secret.env"), "SECRET=***")

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

func TestVersionedCollection(t *testing.T) {
	got := versionedCollection("mudro_docs_current", time.Date(2026, 8, 16, 11, 20, 30, 0, time.UTC))
	if got != "mudro_docs_current_v20260816_112030" {
		t.Fatalf("collection = %q", got)
	}
}

func TestSwitchAliasUsesAtomicBatch(t *testing.T) {
	var method, path string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &body); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := switchAlias(t.Context(), server.URL, "mudro_docs_current", "mudro_docs_current_v20260816_112030"); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || path != "/collections/aliases" {
		t.Fatalf("request = %s %s", method, path)
	}
	actions, ok := body["actions"].([]any)
	if !ok || len(actions) != 2 {
		t.Fatalf("actions = %#v", body["actions"])
	}
	encoded, _ := json.Marshal(body)
	if !strings.Contains(string(encoded), "delete_alias") || !strings.Contains(string(encoded), "create_alias") {
		t.Fatalf("payload = %s", encoded)
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
