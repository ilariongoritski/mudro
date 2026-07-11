package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type qdrantRetriever struct {
	baseURL, collection string
	http                *http.Client
}

func NewQdrantRetrieverFromEnv() (Retriever, error) {
	baseURL := strings.TrimRight(env("QDRANT_URL", "http://qdrant:6333"), "/")
	collection := env("RAG_QDRANT_COLLECTION", "mudro_docs")
	if collection == "" {
		return nil, fmt.Errorf("RAG_QDRANT_COLLECTION is required")
	}
	return &qdrantRetriever{baseURL: baseURL, collection: collection, http: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (r *qdrantRetriever) Search(ctx context.Context, vector []float64, limit int) ([]Source, error) {
	payload, err := json.Marshal(map[string]any{"vector": vector, "limit": limit, "with_payload": true})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/collections/"+r.collection+"/points/search", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query Qdrant: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("query Qdrant: %s", resp.Status)
	}
	var body struct {
		Result []struct {
			Score   float64 `json:"score"`
			Payload struct {
				Path    string `json:"path"`
				Title   string `json:"title"`
				Excerpt string `json:"excerpt"`
			} `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	sources := make([]Source, 0, len(body.Result))
	for _, item := range body.Result {
		if strings.TrimSpace(item.Payload.Path) != "" {
			sources = append(sources, Source{Path: item.Payload.Path, Title: item.Payload.Title, Excerpt: item.Payload.Excerpt, Score: item.Score})
		}
	}
	return sources, nil
}

func qdrantURL() string { return strings.TrimRight(os.Getenv("QDRANT_URL"), "/") }
