package bot

import (
	"errors"
	"strings"
	"testing"
)

func TestTrimMessage(t *testing.T) {
	if got := trimMessage("abc", 2); !strings.Contains(got, "...(truncated)") {
		t.Fatalf("trimmed=%q", got)
	}
	if got := trimMessage("abc", 0); got != "abc" {
		t.Fatalf("limit0=%q", got)
	}
}

func TestFormatReply(t *testing.T) {
	ok := formatReply("/x", []byte("done"), nil)
	if !strings.Contains(ok, "/x:") {
		t.Fatalf("ok=%q", ok)
	}
	fail := formatReply("/x", []byte("ctx"), errors.New("boom"))
	if !strings.Contains(fail, "error: boom") {
		t.Fatalf("fail=%q", fail)
	}
}
