// rag-indexer indexes only approved technical documentation into a versioned Qdrant collection.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/rag"
)

const maxChunkBytes = 1800

type document struct{ Path, Title, Text string }
type point struct {
	ID      string         `json:"id"`
	Vector  []float64      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

func main() {
	ctx := context.Background()
	root := env("MUDRO_ROOT", ".")
	client, err := rag.NewOpenAIClient(rag.OpenAIConfigFromEnv())
	if err != nil {
		fail(err)
	}
	docs, err := collect(root)
	if err != nil {
		fail(err)
	}
	if len(docs) == 0 {
		fail(fmt.Errorf("no approved documentation found"))
	}

	qdrant := strings.TrimRight(env("QDRANT_URL", "http://qdrant:6333"), "/")
	alias := env("RAG_QDRANT_COLLECTION", "mudro_docs_current")
	indexedAt := time.Now().UTC()
	collection := versionedCollection(alias, indexedAt)
	vectors := make([]point, 0, len(docs))
	var dimension int
	for _, doc := range docs {
		vector, embedErr := client.Embed(ctx, doc.Text)
		if embedErr != nil {
			fail(fmt.Errorf("embed %s: %w", doc.Path, embedErr))
		}
		if dimension == 0 {
			dimension = len(vector)
			if ensureErr := ensureCollection(ctx, qdrant, collection, dimension); ensureErr != nil {
				fail(ensureErr)
			}
		}
		contentHash := pointID(doc.Path, doc.Text)
		vectors = append(vectors, point{ID: contentHash, Vector: vector, Payload: map[string]any{
			"path": doc.Path, "title": doc.Title, "excerpt": doc.Text,
			"corpus": "internal_docs", "visibility": "internal",
			"content_hash": contentHash, "indexed_at": indexedAt.Format(time.RFC3339),
		}})
	}
	if err := upsert(ctx, qdrant, collection, vectors); err != nil {
		fail(err)
	}
	if err := switchAlias(ctx, qdrant, alias, collection); err != nil {
		fail(err)
	}
	fmt.Printf("indexed %d documentation chunks into %s via alias %s\n", len(vectors), collection, alias)
}

func versionedCollection(alias string, now time.Time) string {
	return fmt.Sprintf("%s_v%s", alias, now.UTC().Format("20060102_150405"))
}

func collect(root string) ([]document, error) {
	approved := []string{"README.md", "docs", "ops/runbooks", "contracts"}
	var docs []document
	for _, item := range approved {
		path := filepath.Join(root, item)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if !info.IsDir() {
			docs = append(docs, chunksFor(path, item)...)
			continue
		}
		err = filepath.WalkDir(path, func(file string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(file))
			if ext != ".md" && ext != ".yaml" && ext != ".yml" {
				return nil
			}
			relative, relativeErr := filepath.Rel(root, file)
			if relativeErr != nil {
				return relativeErr
			}
			docs = append(docs, chunksFor(file, filepath.ToSlash(relative))...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Path < docs[j].Path })
	return docs, nil
}

func chunksFor(file, relative string) []document {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil
	}
	title := filepath.Base(relative)
	if line := firstHeading(text); line != "" {
		title = line
	}
	var result []document
	for len(text) > 0 {
		end := len(text)
		if end > maxChunkBytes {
			end = strings.LastIndex(text[:maxChunkBytes], "\n")
			if end < maxChunkBytes/2 {
				end = maxChunkBytes
			}
		}
		part := strings.TrimSpace(text[:end])
		if part != "" {
			result = append(result, document{Path: relative, Title: title, Text: part})
		}
		text = strings.TrimSpace(text[end:])
	}
	return result
}

func firstHeading(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	return ""
}

func pointID(path, text string) string {
	sum := sha256.Sum256([]byte(path + "\x00" + text))
	raw := hex.EncodeToString(sum[:16])
	return raw[:8] + "-" + raw[8:12] + "-" + raw[12:16] + "-" + raw[16:20] + "-" + raw[20:32]
}

func ensureCollection(ctx context.Context, base, collection string, dimension int) error {
	url := base + "/collections/" + collection
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("get Qdrant collection: %w", err)
	}
	response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return nil
	}
	if response.StatusCode != http.StatusNotFound {
		return fmt.Errorf("get Qdrant collection: %s", response.Status)
	}
	return request(ctx, http.MethodPut, url, map[string]any{"vectors": map[string]any{"size": dimension, "distance": "Cosine"}})
}

func upsert(ctx context.Context, base, collection string, points []point) error {
	return request(ctx, http.MethodPut, base+"/collections/"+collection+"/points?wait=true", map[string]any{"points": points})
}

func switchAlias(ctx context.Context, base, alias, collection string) error {
	return request(ctx, http.MethodPost, base+"/collections/aliases", map[string]any{"actions": []map[string]any{
		{"delete_alias": map[string]string{"alias_name": alias}},
		{"create_alias": map[string]string{"collection_name": collection, "alias_name": alias}},
	}})
}

func request(ctx context.Context, method, url string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return fmt.Errorf("Qdrant %s: %s", method, response.Status)
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func fail(err error) { fmt.Fprintln(os.Stderr, "rag-indexer:", err); os.Exit(1) }
