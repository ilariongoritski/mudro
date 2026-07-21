package tgexport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLooksLikeMudroAuthor(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		fromID   string
		expected bool
	}{
		{
			name:     "russian mudro in name",
			from:     "Мудро Бот",
			fromID:   "user123",
			expected: true,
		},
		{
			name:     "english mudro in name",
			from:     "Mudro Bot",
			fromID:   "user123",
			expected: true,
		},
		{
			name:     "mudro lowercase in name",
			from:     "some mudro bot",
			fromID:   "user123",
			expected: true,
		},
		{
			name:     "channel with 1001 in name",
			from:     "channel 1001",
			fromID:   "channel1001",
			expected: true,
		},
		{
			name:     "channel with 100.1 in name",
			from:     "channel 100.1",
			fromID:   "channel100.1",
			expected: true,
		},
		{
			name:     "regular user",
			from:     "John Doe",
			fromID:   "user123",
			expected: false,
		},
		{
			name:     "channel without 1001 or 100.1",
			from:     "Some Channel",
			fromID:   "channel999",
			expected: false,
		},
		{
			name:     "empty from",
			from:     "",
			fromID:   "channel1001",
			expected: false,
		},
		{
			name:     "case insensitive channel prefix",
			from:     "Channel 1001",
			fromID:   "CHANNEL1001",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LooksLikeMudroAuthor(tt.from, tt.fromID)
			if result != tt.expected {
				t.Fatalf("LooksLikeMudroAuthor(%q, %q) = %v, want %v", tt.from, tt.fromID, result, tt.expected)
			}
		})
	}
}

func TestIsVisiblePost(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected bool
	}{
		{
			name: "visible post from mudro bot",
			msg: Message{
				Type:             "message",
				ReplyToMessageID: 0,
				From:             "Мудро Бот",
				FromID:           "user123",
			},
			expected: true,
		},
		{
			name: "visible post from channel 1001",
			msg: Message{
				Type:             "message",
				ReplyToMessageID: 0,
				From:             "Channel 1001",
				FromID:           "channel1001",
			},
			expected: true,
		},
		{
			name: "not visible - reply to message",
			msg: Message{
				Type:             "message",
				ReplyToMessageID: 123,
				From:             "Мудро Бот",
				FromID:           "user123",
			},
			expected: false,
		},
		{
			name: "not visible - wrong type",
			msg: Message{
				Type:             "service",
				ReplyToMessageID: 0,
				From:             "Мудро Бот",
				FromID:           "user123",
			},
			expected: false,
		},
		{
			name: "not visible - regular user",
			msg: Message{
				Type:             "message",
				ReplyToMessageID: 0,
				From:             "John Doe",
				FromID:           "user123",
			},
			expected: false,
		},
		{
			name: "not visible - channel without 1001/100.1",
			msg: Message{
				Type:             "message",
				ReplyToMessageID: 0,
				From:             "Some Channel",
				FromID:           "channel999",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVisiblePost(tt.msg)
			if result != tt.expected {
				t.Fatalf("IsVisiblePost(%+v) = %v, want %v", tt.msg, result, tt.expected)
			}
		})
	}
}

func TestNormalizeAuthor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "russian letters",
			input:    "Мудро Бот",
			expected: "мудробот",
		},
		{
			name:     "english letters and spaces",
			input:    "Mudro Bot",
			expected: "mudrobot",
		},
		{
			name:     "with dots, underscores, dashes",
			input:    "Mudro_Bot-123.test",
			expected: "mudrobot123test",
		},
		{
			name:     "with various unicode",
			input:    "Test@#$%Bot",
			expected: "testbot",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "leading/trailing spaces",
			input:    "  Mudro Bot  ",
			expected: "mudrobot",
		},
		{
			name:     "numbers preserved",
			input:    "Channel 1001",
			expected: "channel1001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeAuthor(tt.input)
			if result != tt.expected {
				t.Fatalf("normalizeAuthor(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadVisibleSourcePostIDs(t *testing.T) {
	// Create a temporary JSON export file
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "result.json")

	exportJSON := `{
		"messages": [
			{"id": 1, "type": "message", "from": "Мудро Бот", "from_id": "user1", "reply_to_message_id": 0},
			{"id": 2, "type": "message", "from": "John Doe", "from_id": "user2", "reply_to_message_id": 0},
			{"id": 3, "type": "message", "from": "Мудро Бот", "from_id": "user1", "reply_to_message_id": 5},
			{"id": 4, "type": "service", "from": "Мудро Бот", "from_id": "user1", "reply_to_message_id": 0},
			{"id": 5, "type": "message", "from": "Channel 1001", "from_id": "channel1001", "reply_to_message_id": 0}
		]
	}`

	if err := os.WriteFile(exportPath, []byte(exportJSON), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	ids, err := LoadVisibleSourcePostIDs(exportPath)
	if err != nil {
		t.Fatalf("LoadVisibleSourcePostIDs failed: %v", err)
	}

	expected := []string{"1", "5"}
	if len(ids) != len(expected) {
		t.Fatalf("expected %d IDs, got %d: %v", len(expected), len(ids), ids)
	}
	for i, id := range ids {
		if id != expected[i] {
			t.Fatalf("ids[%d] = %q, want %q", i, id, expected[i])
		}
	}
}

func TestLoadVisibleSourcePostIDs_FileNotFound(t *testing.T) {
	_, err := LoadVisibleSourcePostIDs("/nonexistent/path/result.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadVisibleSourcePostIDs_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "result.json")

	if err := os.WriteFile(exportPath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	_, err := LoadVisibleSourcePostIDs(exportPath)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadVisibleSourcePostIDsFromRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Test case 1: file in data/nu/result.json
	nuDir := filepath.Join(tmpDir, "data", "nu")
	if err := os.MkdirAll(nuDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	exportJSON := `{"messages": [{"id": 42, "type": "message", "from": "Мудро", "from_id": "user1", "reply_to_message_id": 0}]}`
	if err := os.WriteFile(filepath.Join(nuDir, "result.json"), []byte(exportJSON), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	ids, foundPath, err := LoadVisibleSourcePostIDsFromRepo(tmpDir)
	if err != nil {
		t.Fatalf("LoadVisibleSourcePostIDsFromRepo failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != "42" {
		t.Fatalf("expected [\"42\"], got %v", ids)
	}
	if foundPath != filepath.Join(nuDir, "result.json") {
		t.Fatalf("expected path %q, got %q", filepath.Join(nuDir, "result.json"), foundPath)
	}

	// Test case 2: file in data/tg-export/result.json (nu doesn't exist)
	tmpDir2 := t.TempDir()
	tgExportDir := filepath.Join(tmpDir2, "data", "tg-export")
	if err := os.MkdirAll(tgExportDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	exportJSON2 := `{"messages": [{"id": 99, "type": "message", "from": "Mudro", "from_id": "user1", "reply_to_message_id": 0}]}`
	if err := os.WriteFile(filepath.Join(tgExportDir, "result.json"), []byte(exportJSON2), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	ids2, foundPath2, err := LoadVisibleSourcePostIDsFromRepo(tmpDir2)
	if err != nil {
		t.Fatalf("LoadVisibleSourcePostIDsFromRepo failed: %v", err)
	}
	if len(ids2) != 1 || ids2[0] != "99" {
		t.Fatalf("expected [\"99\"], got %v", ids2)
	}
	if foundPath2 != filepath.Join(tgExportDir, "result.json") {
		t.Fatalf("expected path %q, got %q", filepath.Join(tgExportDir, "result.json"), foundPath2)
	}

	// Test case 3: no file found
	tmpDir3 := t.TempDir()
	_, _, err = LoadVisibleSourcePostIDsFromRepo(tmpDir3)
	if err == nil {
		t.Fatal("expected error when no result.json found")
	}
}