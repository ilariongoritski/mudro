package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type todoTask struct {
	Title      string
	StartLine  int
	EndLine    int
	Complexity int
	Action     string
}

func (r *Runner) RunAutoBacklog() ([]byte, error) {
	todoPath := filepath.Join(r.RepoRoot, ".codex", "todo.md")
	donePath := filepath.Join(r.RepoRoot, ".codex", "done.md")

	todoBody, err := os.ReadFile(todoPath)
	if err != nil {
		return nil, fmt.Errorf("read todo: %w", err)
	}
	lines := strings.Split(string(todoBody), "\n")
	tasks := parseTodoTasks(lines)
	if len(tasks) == 0 {
		return []byte("В TODO нет задач для выполнения"), nil
	}

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Complexity == tasks[j].Complexity {
			return tasks[i].Title < tasks[j].Title
		}
		return tasks[i].Complexity < tasks[j].Complexity
	})

	const maxPerRun = 3
	completed := make([]todoTask, 0, maxPerRun)
	skipped := make([]todoTask, 0, len(tasks))
	notes := make([]string, 0, maxPerRun+2)

	for _, t := range tasks {
		if len(completed) >= maxPerRun {
			skipped = append(skipped, t)
			continue
		}
		if t.Action == "" {
			skipped = append(skipped, t)
			continue
		}
		if err := r.executeTaskAction(t.Action); err != nil {
			notes = append(notes, fmt.Sprintf("не выполнено: %s (%s)", t.Title, shortLine(err.Error())))
			skipped = append(skipped, t)
			continue
		}
		completed = append(completed, t)
		notes = append(notes, "выполнено: "+t.Title)
	}

	if len(completed) > 0 {
		if err := rewriteTodoWithout(todoPath, lines, completed); err != nil {
			return nil, err
		}
		if err := appendDone(donePath, completed); err != nil {
			return nil, err
		}
	}

	// После выполнения простых задач — поиск новых важных улучшений.
	findOut, findErr := r.FindImprovements()
	if findErr == nil && len(findOut) > 0 {
		notes = append(notes, "добавлен обзор улучшений через /find")
	}

	var out strings.Builder
	out.WriteString("Работа /rab завершена.\n")
	out.WriteString(fmt.Sprintf("- Выполнено задач: %d\n", len(completed)))
	out.WriteString(fmt.Sprintf("- Отложено/сложно: %d\n", len(skipped)))
	if len(notes) > 0 {
		out.WriteString("Результат:\n")
		for _, n := range notes {
			out.WriteString("- " + n + "\n")
		}
	}
	if len(skipped) > 0 {
		out.WriteString("Нужны уточнения по задачам:\n")
		for i, s := range skipped {
			if i >= 3 {
				out.WriteString("- ...(еще задачи)\n")
				break
			}
			out.WriteString("- " + s.Title + "\n")
		}
	}
	return []byte(strings.TrimSpace(out.String())), nil
}

func parseTodoTasks(lines []string) []todoTask {
	tasks := make([]todoTask, 0, 16)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "- [ ] ") {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(line, "- [ ] "))
		end := i
		for j := i + 1; j < len(lines); j++ {
			next := lines[j]
			if strings.HasPrefix(next, "  - ") || strings.TrimSpace(next) == "" {
				end = j
				continue
			}
			break
		}
		complexity, action := classifyTodoTask(title)
		tasks = append(tasks, todoTask{
			Title:      title,
			StartLine:  i,
			EndLine:    end,
			Complexity: complexity,
			Action:     action,
		})
		i = end
	}
	return tasks
}

func classifyTodoTask(title string) (complexity int, action string) {
	t := strings.ToLower(title)
	switch {
	case strings.Contains(t, "unit-тест"), strings.Contains(t, "unit test"), strings.Contains(t, "тест"):
		return 1, "go_test"
	case strings.Contains(t, "gofmt"), strings.Contains(t, "формат"):
		return 1, "gofmt_repo"
	case strings.Contains(t, "find"), strings.Contains(t, "улучш"):
		return 2, "find_scan"
	default:
		return 9, ""
	}
}

func (r *Runner) executeTaskAction(action string) error {
	switch action {
	case "go_test":
		_, err := r.runStep([]string{"go", "test", "./..."})
		return err
	case "gofmt_repo":
		list, err := r.runStep([]string{"git", "ls-files", "*.go"})
		if err != nil {
			return err
		}
		files := strings.Fields(string(list))
		if len(files) == 0 {
			return nil
		}
		args := append([]string{"-w"}, files...)
		_, err = r.runStep(append([]string{"gofmt"}, args...))
		return err
	case "find_scan":
		_, err := r.FindImprovements()
		return err
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
}

func rewriteTodoWithout(todoPath string, lines []string, completed []todoTask) error {
	sort.Slice(completed, func(i, j int) bool { return completed[i].StartLine > completed[j].StartLine })
	for _, c := range completed {
		if c.StartLine < 0 || c.EndLine >= len(lines) || c.StartLine > c.EndLine {
			continue
		}
		lines = append(lines[:c.StartLine], lines[c.EndLine+1:]...)
	}
	out := strings.Join(lines, "\n")
	return os.WriteFile(todoPath, []byte(out), 0o644)
}

func appendDone(donePath string, completed []todoTask) error {
	b, err := os.ReadFile(donePath)
	if err != nil {
		return fmt.Errorf("read done: %w", err)
	}
	body := string(b)
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	date := time.Now().Format("2006-01-02")
	for _, c := range completed {
		body += fmt.Sprintf("- %s | %s | Эффект: выполнено автоматикой `/rab`\n", date, c.Title)
	}
	return os.WriteFile(donePath, []byte(body), 0o644)
}
