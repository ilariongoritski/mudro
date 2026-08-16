package rag

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var ErrInsufficientContext = errors.New("insufficient documentation context")

type Service struct {
	embedder  Embedder
	retriever Retriever
	generator Generator
}

func NewService(embedder Embedder, retriever Retriever, generator Generator) *Service {
	return &Service{embedder: embedder, retriever: retriever, generator: generator}
}

func (s *Service) Ask(ctx context.Context, question string) (Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return Answer{}, errors.New("question is required")
	}

	vector, err := s.embedder.Embed(ctx, question)
	if err != nil {
		return Answer{}, err
	}
	sources, err := s.retriever.Search(ctx, vector, 5)
	if err != nil {
		return Answer{}, err
	}
	sources = relevantSources(sources, minimumScore())
	if len(sources) == 0 {
		return Answer{Sources: []Source{}}, ErrInsufficientContext
	}

	answer, err := s.generator.Generate(ctx, question, sources)
	if err != nil {
		return Answer{}, err
	}
	return Answer{Answer: answer, Sources: sources, Grounded: true}, nil
}

func (s *Service) Ready(ctx context.Context) error {
	if err := s.retriever.Ready(ctx); err != nil {
		return fmt.Errorf("RAG retriever unavailable: %w", err)
	}
	return nil
}

func relevantSources(sources []Source, minScore float64) []Source {
	filtered := make([]Source, 0, len(sources))
	for _, source := range sources {
		if source.Score >= minScore {
			filtered = append(filtered, source)
		}
	}
	return filtered
}

func minimumScore() float64 {
	value := strings.TrimSpace(os.Getenv("RAG_MIN_SCORE"))
	if value == "" {
		return 0.65
	}
	score, err := strconv.ParseFloat(value, 64)
	if err != nil || score < 0 || score > 1 {
		return 0.65
	}
	return score
}
