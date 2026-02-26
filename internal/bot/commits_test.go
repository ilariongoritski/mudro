package bot

import "testing"

func TestParseCommitSummaries(t *testing.T) {
	raw := `__C__abc123|2026-02-25|fix parser
M	internal/bot/handler.go
A	README.md
__C__def456|2026-02-24|docs
D	internal/config/config.go`

	got := parseCommitSummaries(raw)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Hash != "abc123" || got[0].Added != 1 || got[0].Modified != 1 || got[0].Deleted != 0 {
		t.Fatalf("unexpected first commit stats: %+v", got[0])
	}
	if got[1].Deleted != 1 {
		t.Fatalf("unexpected second commit stats: %+v", got[1])
	}
}

func TestClassifyDomain(t *testing.T) {
	cases := map[string]string{
		"internal/bot/handler.go": "бот/команды",
		"cmd/bot/main.go":         "запуск бота",
		"README.md":               "документация",
		"Makefile":                "инфраструктура",
		"internal/agent/repo.go":  "internal",
	}
	for in, want := range cases {
		if got := classifyDomain(in); got != want {
			t.Fatalf("classifyDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestShortEssence(t *testing.T) {
	c := commitSummary{
		Subject: "add tests",
		Domains: map[string]int{"бот/команды": 2, "документация": 1},
	}
	got := c.shortEssence()
	if got == "" || got[:1] != "A" {
		t.Fatalf("shortEssence = %q, want capitalized", got)
	}
}
