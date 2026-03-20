package app

import (
	"encoding/json"
	"testing"
)

func TestBuildMediaJSONAssignsUniqueTrackPositionAfterMedia(t *testing.T) {
	payload, err := buildMediaJSON(
		[]mediaItem{
			{
				Kind:     "photo",
				URL:      "photos/a.jpg",
				Position: 1,
			},
		},
		[]string{"Track One"},
	)
	if err != nil {
		t.Fatalf("buildMediaJSON returned error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len(decoded) = %d, want 2", len(decoded))
	}

	if got := int(decoded[0]["position"].(float64)); got != 1 {
		t.Fatalf("media position = %d, want 1", got)
	}
	if got := int(decoded[1]["position"].(float64)); got != 2 {
		t.Fatalf("track position = %d, want 2", got)
	}
}

func TestBuildMediaJSONAppendsTracksAfterHighestExistingPosition(t *testing.T) {
	payload, err := buildMediaJSON(
		[]mediaItem{
			{
				Kind:     "photo",
				URL:      "photos/a.jpg",
				Position: 10,
			},
		},
		[]string{"Track One"},
	)
	if err != nil {
		t.Fatalf("buildMediaJSON returned error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len(decoded) = %d, want 2", len(decoded))
	}

	if got := int(decoded[1]["position"].(float64)); got != 11 {
		t.Fatalf("track position = %d, want 11", got)
	}
}
