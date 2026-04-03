package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/agent/domain"
)

// makeTargetRe matches "make <target>" patterns in todo text.
var makeTargetRe = regexp.MustCompile(`(?i)\bmake\s+([\w][\w\-]*)`)

type Worker struct {
	RepoRoot string
	Usecase  TaskUsecase
	WorkerID string
}

func (w *Worker) RunOnce(ctx context.Context) (bool, error) {
	task, err := w.Usecase.ClaimNext(ctx, w.WorkerID)
	if err != nil {
		return false, err
	}
	if task == nil {
		return false, nil
	}

	err = w.processTask(ctx, task)
	if err == nil {
		return true, w.Usecase.CompleteTask(ctx, task.ID)
	}

	retryAfter := time.Duration(0)
	if task.Attempts < task.MaxAttempts {
		retryAfter = 2 * time.Minute
	}
	markErr := w.Usecase.FailTask(ctx, task.ID, err.Error(), retryAfter)
	if markErr != nil {
		return true, errors.Join(err, markErr)
	}
	return true, err
}

func (w *Worker) processTask(ctx context.Context, task *domain.Task) error {
	switch task.Kind {
	case "todo_item":
		return w.processTodoTask(ctx, task)
	case "health_check":
		return w.runCommand(ctx, "make", "test")
	case "db_check":
		return w.runCommand(ctx, "make", "dbcheck")
	case "tables_check":
		return w.runCommand(ctx, "make", "tables")
	case "count_posts":
		return w.runCommand(ctx, "make", "count-posts")
	default:
		return fmt.Errorf("unsupported task kind: %s", task.Kind)
	}
}

func (w *Worker) processTodoTask(ctx context.Context, task *domain.Task) error {
	var payload struct {
		Source string `json:"source"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	text := strings.TrimSpace(payload.Text)
	if text == "" {
		return errors.New("empty todo text")
	}

	runID := time.Now().Format("20060102-150405") + fmt.Sprintf("-task-%d", task.ID)
	logDir := filepath.Join(w.RepoRoot, ".codex", "logs", runID)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("mkdir log dir: %w", err)
	}

	// Detect and run "make <target>" mentioned in the todo text.
	var cmdOutput string
	var cmdErr error
	var cmdLabel string

	if match := makeTargetRe.FindStringSubmatch(text); match != nil {
		target := match[1]
		cmdLabel = "make " + target
		taskCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		cmdOutput, cmdErr = w.runCommandOutput(taskCtx, "make", target)
	}

	status := "зафиксировано (ручное действие)"
	if cmdLabel != "" && cmdErr == nil {
		status = "выполнено: " + cmdLabel
	} else if cmdLabel != "" && cmdErr != nil {
		status = "ошибка при выполнении " + cmdLabel + ": " + cmdErr.Error()
	}

	lines := []string{
		"# agent worker — todo task",
		"",
		"- Дата/время: " + time.Now().Format(time.RFC3339),
		"- Worker:    " + w.WorkerID,
		fmt.Sprintf("- Task ID:   %d", task.ID),
		"- Source:    " + payload.Source,
		"- Text:      " + text,
		"- Статус:    " + status,
	}
	if cmdOutput != "" {
		lines = append(lines, "", "## Вывод команды", "", "```", cmdOutput, "```")
	}

	logText := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(logDir, "agent-task.md"), []byte(logText), 0o644); err != nil {
		return fmt.Errorf("write task log: %w", err)
	}

	return cmdErr
}

// runCommand runs a command and returns an error if it fails.
func (w *Worker) runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = w.RepoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %w\n%s", name, args, err, string(out))
	}
	return nil
}

// runCommandOutput runs a command and returns combined stdout+stderr as a string.
func (w *Worker) runCommandOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = w.RepoRoot
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return output, fmt.Errorf("%s %v: %w", name, args, err)
	}
	return output, nil
}
