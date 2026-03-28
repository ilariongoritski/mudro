package media

import (
	"encoding/json"
	"testing"
)

func TestParseLegacyJSONPreservesMediaFields(t *testing.T) {
	raw := json.RawMessage(`[
	  {"kind":"photo","url":"photos/a.jpg","width":1280,"height":720,"position":2,"extra":{"file_name":"a.jpg","mime_type":"image/jpeg"}},
	  {"kind":"gif","url":"stickers/s.webp","preview_url":"stickers/s_thumb.jpg","extra":{"media_type":"sticker"}},
	  {"kind":"audio","title":"Voice note"}
	]`)

	items := ParseLegacyJSON(raw)
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}

	if items[0].Title != "a.jpg" || items[0].Width != 1280 || items[0].Height != 720 || items[0].Position != 2 {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].PreviewURL != "stickers/s_thumb.jpg" || items[1].Position != 3 {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
	if items[2].Title != "Voice note" || items[2].Position != 4 {
		t.Fatalf("unexpected third item: %+v", items[2])
	}
}

func TestNormalizeJSONProducesCanonicalShape(t *testing.T) {
	raw := json.RawMessage(`[
	  {"Kind":"photo","URL":"photos/a.jpg","Extra":{"file_name":"a.jpg"}}
	]`)

	normalized := NormalizeJSON(raw)
	if len(normalized) == 0 {
		t.Fatal("NormalizeJSON returned empty payload")
	}

	var decoded []Item
	if err := json.Unmarshal(normalized, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Kind != "photo" || decoded[0].Title != "a.jpg" {
		t.Fatalf("unexpected normalized payload: %+v", decoded)
	}
}

func TestAssetKeyStableForSameAsset(t *testing.T) {
	itemA := Item{
		Kind:       "photo",
		URL:        "photos/a.jpg",
		PreviewURL: "photos/a_thumb.jpg",
		Title:      "a.jpg",
		Width:      800,
		Height:     600,
		Position:   1,
		Extra:      map[string]any{"mime_type": "image/jpeg", "file_name": "a.jpg"},
	}
	itemB := itemA
	itemB.Position = 99

	keyA, err := assetKey("tg", itemA)
	if err != nil {
		t.Fatalf("assetKey(itemA): %v", err)
	}
	keyB, err := assetKey("tg", itemB)
	if err != nil {
		t.Fatalf("assetKey(itemB): %v", err)
	}
	if keyA != keyB {
		t.Fatalf("asset key should ignore position: %s != %s", keyA, keyB)
	}
}

func TestParseLegacyJSONAssignsUniqueAutoPositionsAfterExplicitOffset(t *testing.T) {
	raw := json.RawMessage(`[
	  {"kind":"photo","url":"photos/a.jpg","position":2},
	  {"kind":"photo","url":"photos/b.jpg","position":2},
	  {"kind":"photo","url":"photos/c.jpg"}
	]`)

	items := ParseLegacyJSON(raw)
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}

	if items[0].Position != 2 {
		t.Fatalf("items[0].Position = %d, want 2", items[0].Position)
	}
	if items[1].Position != 3 {
		t.Fatalf("items[1].Position = %d, want 3", items[1].Position)
	}
	if items[2].Position != 4 {
		t.Fatalf("items[2].Position = %d, want 4", items[2].Position)
	}
}

func TestParseLegacyJSONSkipsEmptyRows(t *testing.T) {
	raw := json.RawMessage(`[
	  {},
	  {"kind":"photo","url":"photos/a.jpg"}
	]`)

	items := ParseLegacyJSON(raw)
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Kind != "photo" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}
