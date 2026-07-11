package rag

import (
	"context"
	"errors"
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
	if len(sources) == 0 {
		return Answer{Sources: []Source{}}, ErrInsufficientContext
	}

	answer, err := s.generator.Generate(ctx, question, sources)
	if err != nil {
		return Answer{}, err
	}
	return Answer{Answer: answer, Sources: sources, Grounded: true}, nil
}
