package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goritskimihail/mudro/internal/agent/domain"
)

type mockUsecase struct {
	claimNextFn      func(ctx context.Context, workerID string) (*domain.Task, error)
	completeTaskFn   func(ctx context.Context, taskID int64) error
	failTaskFn       func(ctx context.Context, taskID int64, errText string, retryAfter time.Duration) error
	enqueueFn        func(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error)
	enqueueWaitingFn func(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error)
	approveTaskFn    func(ctx context.Context, taskID int64) error
	rejectTaskFn     func(ctx context.Context, taskID int64, reason string) error
}

func (m *mockUsecase) ClaimNext(ctx context.Context, workerID string) (*domain.Task, error) {
	if m.claimNextFn != nil {
		return m.claimNextFn(ctx, workerID)
	}
	return nil, nil
}

func (m *mockUsecase) CompleteTask(ctx context.Context, taskID int64) error {
	if m.completeTaskFn != nil {
		return m.completeTaskFn(ctx, taskID)
	}
	return nil
}

func (m *mockUsecase) FailTask(ctx context.Context, taskID int64, errText string, retryAfter time.Duration) error {
	if m.failTaskFn != nil {
		return m.failTaskFn(ctx, taskID, errText, retryAfter)
	}
	return nil
}

func (m *mockUsecase) Enqueue(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	if m.enqueueFn != nil {
		return m.enqueueFn(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey)
	}
	return 1, nil
}

func (m *mockUsecase) EnqueueWaitingApproval(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
	if m.enqueueWaitingFn != nil {
		return m.enqueueWaitingFn(ctx, kind, payload, priority, runAfter, maxAttempts, dedupeKey)
	}
	return 2, nil
}

func (m *mockUsecase) ApproveTask(ctx context.Context, taskID int64) error {
	if m.approveTaskFn != nil {
		return m.approveTaskFn(ctx, taskID)
	}
	return nil
}

func (m *mockUsecase) RejectTask(ctx context.Context, taskID int64, reason string) error {
	if m.rejectTaskFn != nil {
		return m.rejectTaskFn(ctx, taskID, reason)
	}
	return nil
}

func TestIsRiskyTodo(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"drop table", "DROP TABLE users", true},
		{"truncate table", "TRUNCATE TABLE logs", true},
		{"reset database", "RESET DATABASE", true},
		{"rm -rf", "rm -rf /tmp/data", true},
		{"docker compose down -v", "docker compose down -v", true},
		{"alter table", "ALTER TABLE posts ADD COLUMN x int", true},
		{"delete from", "DELETE FROM posts WHERE id=1", true},
		{"case insensitive drop", "drop table users", true},
		{"case insensitive rm", "RM -RF /", true},
		{"safe todo", "fix login bug", false},
		{"safe with make", "make test", false},
		{"safe with dbcheck", "make dbcheck", false},
		{"empty", "", false},
		{"whitespace only", "   ", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isRiskyTodo(tc.text)
			if got != tc.want {
				t.Errorf("isRiskyTodo(%q) = %v, want %v", tc.text, got, tc.want)
			}
		})
	}
}

func TestDetectTaskKind(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"dbcheck keyword", "make dbcheck", "db_check"},
		{"dbcheck text", "проверка бд", "db_check"},
		{"dbcheck plural", "проверки бд", "db_check"},
		{"select 1", "SELECT 1", "db_check"},
		{"tables list", "make tables", "tables_check"},
		{"tables russian", "список таблиц", "tables_check"},
		{"dt command", "\\dt", "tables_check"},
		{"count posts", "make count-posts", "count_posts"},
		{"count sql", "count(*) from posts", "count_posts"},
		{"count russian", "количество постов", "count_posts"},
		{"health check make test", "make test", "health_check"},
		{"health check go test", "go test ./...", "health_check"},
		{"health check text", "health check", "health_check"},
		{"health check russian", "запустить тест", "health_check"},
		{"default todo", "refactor auth module", "todo_item"},
		{"default empty", "something random", "todo_item"},
		{"case insensitive", "MAKE DBCHECK", "db_check"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectTaskKind(tc.text)
			if got != tc.want {
				t.Errorf("detectTaskKind(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}

func TestPlanFromTodo(t *testing.T) {
	tests := []struct {
		name        string
		todoContent string
		wantCount   int
		wantKinds   []string
		wantErr     bool
	}{
		{
			name:        "empty file",
			todoContent: "",
			wantCount:   0,
			wantKinds:   nil,
			wantErr:     false,
		},
		{
			name:        "single safe todo",
			todoContent: "- [ ] fix login bug\n",
			wantCount:   1,
			wantKinds:   []string{"todo_item"},
			wantErr:     false,
		},
		{
			name:        "single risky todo",
			todoContent: "- [ ] DROP TABLE users\n",
			wantCount:   1,
			wantKinds:   []string{"todo_item"},
			wantErr:     false,
		},
		{
			name:        "multiple todos mixed",
			todoContent: "- [ ] fix login bug\n- [ ] make dbcheck\n- [ ] DROP TABLE users\n- [ ] make tables\n",
			wantCount:   4,
			wantKinds:   []string{"todo_item", "db_check", "todo_item", "tables_check"},
			wantErr:     false,
		},
		{
			name:        "ignores non-checkbox lines",
			todoContent: "# Header\n- [x] done item\n- [ ] todo item\nplain text\n",
			wantCount:   1,
			wantKinds:   []string{"todo_item"},
			wantErr:     false,
		},
		{
			name:        "ignores checked items",
			todoContent: "- [x] done\n- [ ] pending\n",
			wantCount:   1,
			wantKinds:   []string{"todo_item"},
			wantErr:     false,
		},
		{
			name:        "whitespace handling",
			todoContent: "  - [ ]  task with spaces  \n  - [ ]  tabbed\n",
			wantCount:   2,
			wantKinds:   []string{"todo_item", "todo_item"},
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			codexDir := filepath.Join(tmpDir, ".codex")
			if err := os.MkdirAll(codexDir, 0755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			todoPath := filepath.Join(codexDir, "todo.md")
			if err := os.WriteFile(todoPath, []byte(tc.todoContent), 0644); err != nil {
				t.Fatalf("write todo: %v", err)
			}

			mock := &mockUsecase{
				enqueueFn: func(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
					return 1, nil
				},
				enqueueWaitingFn: func(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
					return 2, nil
				},
			}

			count, err := PlanFromTodo(context.Background(), tmpDir, mock)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("PlanFromTodo error: %v", err)
			}
			if count != tc.wantCount {
				t.Errorf("PlanFromTodo count = %d, want %d", count, tc.wantCount)
			}
		})
	}
}

func TestPlanFromTodo_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create .codex/todo.md

	mock := &mockUsecase{}
	count, err := PlanFromTodo(context.Background(), tmpDir, mock)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for missing file, got %d", count)
	}
}

func TestPlanFromTodo_EnqueueError(t *testing.T) {
	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	todoPath := filepath.Join(codexDir, "todo.md")
	if err := os.WriteFile(todoPath, []byte("- [ ] task\n"), 0644); err != nil {
		t.Fatalf("write todo: %v", err)
	}

	mock := &mockUsecase{
		enqueueFn: func(ctx context.Context, kind string, payload any, priority int, runAfter time.Time, maxAttempts int, dedupeKey string) (int64, error) {
			return 0, errors.New("enqueue failed")
		},
	}

	_, err := PlanFromTodo(context.Background(), tmpDir, mock)
	if err == nil {
		t.Error("expected error on enqueue failure")
	}
}

func TestWorker_RunCommandOutput(t *testing.T) {
	tmpDir := t.TempDir()
	w := &Worker{RepoRoot: tmpDir, WorkerID: "test-worker"}

	// Test successful command
	output, err := w.runCommandOutput(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("runCommandOutput error: %v", err)
	}
	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}

	// Test failing command
	_, err = w.runCommandOutput(context.Background(), "false")
	if err == nil {
		t.Error("expected error for failing command")
	}

	// Test command not found
	_, err = w.runCommandOutput(context.Background(), "nonexistent_command_xyz")
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestWorker_ProcessTodoTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal Makefile for make command tests
	makefileContent := `help:
	@echo "help target"
test:
	@echo "test target"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	tests := []struct {
		name         string
		payload      string
		setupMock    func(*mockUsecase)
		wantErr      bool
		checkLog     func(t *testing.T, logDir string)
	}{
		{
			name:    "empty text",
			payload: `{"source":"todo.md","text":""}`,
			setupMock: func(m *mockUsecase) {},
			wantErr: true,
			checkLog: nil,
		},
		{
			name:    "simple todo",
			payload: `{"source":"todo.md","text":"fix login bug"}`,
			setupMock: func(m *mockUsecase) {},
			wantErr: false,
			checkLog: func(t *testing.T, logDir string) {
				files, _ := os.ReadDir(logDir)
				if len(files) == 0 {
					t.Error("expected log file to be created")
				}
			},
		},
		{
			name:    "todo with make target",
			payload: `{"source":"todo.md","text":"make test"}`,
			setupMock: func(m *mockUsecase) {},
			wantErr: false,
			checkLog: func(t *testing.T, logDir string) {
				files, _ := os.ReadDir(logDir)
				if len(files) == 0 {
					t.Error("expected log file")
					return
				}
				content, _ := os.ReadFile(filepath.Join(logDir, files[0].Name()))
				if len(content) == 0 {
					t.Error("log should not be empty")
				}
			},
		},
		{
			name:    "todo with make target - success",
			payload: `{"source":"todo.md","text":"make help"}`,
			setupMock: func(m *mockUsecase) {},
			wantErr: false,
			checkLog: func(t *testing.T, logDir string) {
				files, _ := os.ReadDir(logDir)
				if len(files) == 0 {
					t.Error("expected log file")
					return
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockUsecase{}
			tc.setupMock(mock)

			w := &Worker{
				RepoRoot: tmpDir,
				Usecase:  mock,
				WorkerID: "test-worker",
			}

			task := &domain.Task{
				ID:      1,
				Kind:    "todo_item",
				Payload: []byte(tc.payload),
			}

			err := w.processTodoTask(context.Background(), task)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("processTodoTask error: %v", err)
			}

			if tc.checkLog != nil {
				codexDir := filepath.Join(tmpDir, ".codex", "logs")
				files, _ := os.ReadDir(codexDir)
				if len(files) > 0 {
					latestLog := files[len(files)-1].Name()
					logDir := filepath.Join(codexDir, latestLog)
					tc.checkLog(t, logDir)
				}
			}
		})
	}
}

func TestWorker_RunOnce_NoTask(t *testing.T) {
	mock := &mockUsecase{
		claimNextFn: func(ctx context.Context, workerID string) (*domain.Task, error) {
			return nil, nil // no task available
		},
	}

	w := &Worker{
		RepoRoot: t.TempDir(),
		Usecase:  mock,
		WorkerID: "test-worker",
	}

	processed, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce error: %v", err)
	}
	if processed {
		t.Error("expected false when no task")
	}
}