package bot

import (
	"testing"
	"time"
)

func TestEstimateRunSeconds(t *testing.T) {
	runs := []runRef{
		{id: "a", ts: time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)},
		{id: "b", ts: time.Date(2026, 2, 26, 10, 1, 0, 0, time.UTC)},
		{id: "c", ts: time.Date(2026, 2, 26, 11, 30, 0, 0, time.UTC)},
	}
	if got := estimateRunSeconds(runs[0], runs, 0); got != 120 {
		t.Fatalf("min clamp got=%d", got)
	}
	if got := estimateRunSeconds(runs[1], runs, 1); got != 600 {
		t.Fatalf("max clamp to default got=%d", got)
	}
	if got := estimateRunSeconds(runs[2], runs, 2); got != 600 {
		t.Fatalf("last run default got=%d", got)
	}
}

func TestTimeFormattingHelpers(t *testing.T) {
	if got := fmtDuration(3661); got != "01:01:01" {
		t.Fatalf("fmtDuration=%q", got)
	}
	if got := formatHoursMinutes(3661); got != "1 часов 1 минут" {
		t.Fatalf("formatHoursMinutes=%q", got)
	}
	if got := estimateEvaporationML(3600); got != 50 {
		t.Fatalf("estimateEvaporationML=%d", got)
	}
}

func TestTopRuntimeCommands(t *testing.T) {
	rt := &runtimeTimeMemory{
		ByCommand: map[string]runtimeCommand{
			"a": {TotalMS: 10, Responses: 1},
			"b": {TotalMS: 30, Responses: 2},
		},
	}
	out := topRuntimeCommands(rt, 1)
	if len(out) != 1 || out[0][:2] != "/b" {
		t.Fatalf("unexpected top commands: %#v", out)
	}
}
