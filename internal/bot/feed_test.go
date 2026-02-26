package bot

import (
	"strings"
	"testing"
)

func TestShortText(t *testing.T) {
	if got := shortText(nil); got != "(без текста)" {
		t.Fatalf("nil shortText=%q", got)
	}
	s := "  "
	if got := shortText(&s); got != "(без текста)" {
		t.Fatalf("empty shortText=%q", got)
	}
	long := strings.Repeat("a", 90)
	got := shortText(&long)
	if len([]rune(got)) != 83 || !strings.HasSuffix(got, "...") {
		t.Fatalf("long shortText=%q", got)
	}
}
