package rag

import "context"

// Embedder converts text into a fixed-size semantic vector.
type Embedder interface {
	Embed(context.Context, string) ([]float64, error)
}

// Retriever returns the most relevant indexed documentation chunks.
type Retriever interface {
	Search(context.Context, []float64, int) ([]Source, error)
}

// Generator creates an answer using only supplied documentation excerpts.
type Generator interface {
	Generate(context.Context, string, []Source) (string, error)
}

type Source struct {
	Path    string  `json:"path"`
	Title   string  `json:"title"`
	Excerpt string  `json:"excerpt"`
	Score   float64 `json:"score"`
}

type Answer struct {
	Answer   string   `json:"answer"`
	Sources  []Source `json:"sources"`
	Grounded bool     `json:"grounded"`
}
