package tgexport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLooksLikeMudroAuthor(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		fromID  string
		want    bool
	}{
		{"cyrillic mudro", "Мудро", "123", true},
		{"latin mudro", "Mudro", "123", true},
		{"mixed case", "MuDrO", "123", true},
		{"with prefix", "Mudro Bot", "123", true},
		{"channel 1001", "Some Channel", "channel1001", true},
		{"channel 100.1", "Some Channel", "channel100.1", true},
		{"regular user", "User", "123", false},
		{"empty", "", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := LooksLikeMudroAuthor(tc.from, tc.fromID)
			if got != tc.want {
				t.Errorf("LooksLikeMudroAuthor(%q, %q) = %v, want %v", tc.from, tc.fromID, got, tc.want)
			}
		})
	}
}

func TestIsVisiblePost(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		want    bool
	}{
		{"visible post", Message{Type: "message", ReplyToMessageID: 0, From: "Mudro", FromID: "123"}, true},
		{"reply not visible", Message{Type: "message", ReplyToMessageID: 456, From: "Mudro", FromID: "123"}, false},
		{"non-message type", Message{Type: "service", ReplyToMessageID: 0, From: "Mudro", FromID: "123"}, false},
		{"not mudro author", Message{Type: "message", ReplyToMessageID: 0, From: "User", FromID: "123"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsVisiblePost(tc.msg)
			if got != tc.want {
				t.Errorf("IsVisiblePost(%+v) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestLoadVisibleSourcePostIDs(t *testing.T) {
	// Create temp file with sample export
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.json")

	content := `{
		"messages": [
			{"id": 1, "type": "message", "from": "Mudro", "from_id": "123", "reply_to_message_id": 0},
			{"id": 2, "type": "message", "from": "User", "from_id": "456", "reply_to_message_id": 0},
			{"id": 3, "type": "message", "from": "Mudro", "from_id": "123", "reply_to_message_id": 2},
			{"id": 4, "type": "service", "from": "Mudro", "from_id": "123", "reply_to_message_id": 0}
		]
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	ids, err := LoadVisibleSourcePostIDs(path)
	if err != nil {
		t.Fatalf("LoadVisibleSourcePostIDs: %v", err)
	}

	if len(ids) != 1 || ids[0] != "1" {
		t.Errorf("expected [\"1\"], got %v", ids)
	}
}

func TestLoadVisibleSourcePostIDsFromRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file in data/nu/
	nuDir := filepath.Join(tmpDir, "data", "nu")
	if err := os.MkdirAll(nuDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := `{"messages": [{"id": 42, "type": "message", "from": "Mudro", "from_id": "123", "reply_to_message_id": 0}]}`
	if err := os.WriteFile(filepath.Join(nuDir, "result.json"), []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ids, usedPath, err := LoadVisibleSourcePostIDsFromRepo(tmpDir)
	if err != nil {
		t.Fatalf("LoadVisibleSourcePostIDsFromRepo: %v", err)
	}

	if len(ids) != 1 || ids[0] != "42" {
		t.Errorf("expected [\"42\"], got %v", ids)
	}
	if usedPath != filepath.Join(nuDir, "result.json") {
		t.Errorf("unexpected path: %s", usedPath)
	}
}

func TestLoadVisibleSourcePostIDsFromRepoNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, _, err := LoadVisibleSourcePostIDsFromRepo(tmpDir)
	if err == nil {
		t.Error("expected error when file not found")
	}
}

func TestNormalizeAuthor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "Mudro", "mudro"},
		{"with spaces", "  Mudro Bot  ", "mudrobot"},
		{"with punctuation", "Mudro-Bot_v2.1", "mudrobotv21"},
		{"cyrillic", "Мудро", "мудро"},
		{"empty", "", ""},
		{"only spaces", "   ", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeAuthor(tc.in)
			if got != tc.want {
				t.Errorf("normalizeAuthor(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}