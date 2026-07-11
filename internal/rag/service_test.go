package rag

import (
	"context"
	"testing"
)

type testEmbedder struct{}

func (testEmbedder) Embed(context.Context, string) ([]float64, error) { return []float64{1, 2}, nil }

type testRetriever struct{ sources []Source }

func (r testRetriever) Search(context.Context, []float64, int) ([]Source, error) {
	return r.sources, nil
}

type testGenerator struct{}

func (testGenerator) Generate(context.Context, string, []Source) (string, error) {
	return "grounded answer", nil
}
func TestAskRequiresSources(t *testing.T) {
	_, err := NewService(testEmbedder{}, testRetriever{}, testGenerator{}).Ask(context.Background(), "question")
	if err != ErrInsufficientContext {
		t.Fatalf("err = %v", err)
	}
}
func TestAskReturnsGroundedAnswer(t *testing.T) {
	answer, err := NewService(testEmbedder{}, testRetriever{sources: []Source{{Path: "docs/a.md", Excerpt: "evidence"}}}, testGenerator{}).Ask(context.Background(), "question")
	if err != nil || !answer.Grounded || answer.Answer != "grounded answer" {
		t.Fatalf("answer=%+v err=%v", answer, err)
	}
}
