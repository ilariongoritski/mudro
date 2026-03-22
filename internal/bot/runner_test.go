package bot

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func shellWriteCmd(s string) []string {
	if runtime.GOOS == "windows" {
		return []string{"powershell", "-NoProfile", "-Command", "[Console]::Out.Write('" + s + "')"}
	}
	return []string{"sh", "-c", "printf " + s}
}

func TestRunnerRunStepAndRunSteps(t *testing.T) {
	r := &Runner{RepoRoot: ".", Timeout: 5 * time.Second}

	out, err := r.runStep(shellWriteCmd("hi"))
	if err != nil {
		t.Fatalf("runStep: %v", err)
	}
	if string(out) != "hi" {
		t.Fatalf("runStep output=%q", string(out))
	}

	if _, err := r.runStep(nil); err == nil {
		t.Fatal("expected error for empty step")
	}

	stepOne := shellWriteCmd("one")
	stepTwo := shellWriteCmd("two")
	out, err = r.runSteps([][]string{stepOne, stepTwo})
	if err != nil {
		t.Fatalf("runSteps: %v", err)
	}
	if !strings.Contains(string(out), "$ "+strings.Join(stepOne, " ")) {
		t.Fatalf("unexpected runSteps output: %q", string(out))
	}
}
