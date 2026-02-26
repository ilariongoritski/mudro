package bot

import "testing"

func TestParseShortstat(t *testing.T) {
	in := " 2 files changed, 10 insertions(+), 3 deletions(-)\n1 file changed, 1 insertion(+)"
	files, ins, del := parseShortstat(in)
	if files != 3 || ins != 11 || del != 3 {
		t.Fatalf("got files=%d ins=%d del=%d", files, ins, del)
	}
}

func TestLeadingInt(t *testing.T) {
	if got := leadingInt("12 files changed"); got != 12 {
		t.Fatalf("got=%d", got)
	}
	if got := leadingInt("x files changed"); got != -1 {
		t.Fatalf("got=%d", got)
	}
}
