package bot

import (
	"strings"
	"testing"
	"time"
)

func TestRunnerRunStepAndRunSteps(t *testing.T) {
	r := &Runner{RepoRoot: ".", Timeout: 5 * time.Second}

	out, err := r.runStep([]string{"sh", "-c", "printf hi"})
	if err != nil {
		t.Fatalf("runStep: %v", err)
	}
	if string(out) != "hi" {
		t.Fatalf("runStep output=%q", string(out))
	}

	if _, err := r.runStep(nil); err == nil {
		t.Fatal("expected error for empty step")
	}

	out, err = r.runSteps([][]string{{"sh", "-c", "printf one"}, {"sh", "-c", "printf two"}})
	if err != nil {
		t.Fatalf("runSteps: %v", err)
	}
	if !strings.Contains(string(out), "$ sh -c printf one") {
		t.Fatalf("unexpected runSteps output: %q", string(out))
	}
}
