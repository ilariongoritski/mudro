package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

func (r *Runner) HealthDaily() ([]byte, error) {
	now := time.Now()
	var out strings.Builder

	psOut, psErr := r.runStep([]string{"docker", "compose", "ps", "db"})
	dbOut, dbErr := r.runStep([]string{"make", "dbcheck"})

	health := "degraded"
	if psErr == nil && dbErr == nil {
		health = "ok"
	}
	out.WriteString("Состояние сейчас: " + health + "\n")
	if psErr != nil {
		out.WriteString("db container: error: " + psErr.Error() + "\n")
	} else {
		out.WriteString("db container: ok\n")
	}
	if dbErr != nil {
		out.WriteString("dbcheck: error: " + dbErr.Error() + "\n")
	} else {
		out.WriteString("dbcheck: ok\n")
	}

	logSummary, logErr := r.summarizeTodayRuns(now)
	out.WriteString("\nЗа день (логи):\n")
	if logErr != nil {
		out.WriteString("не удалось собрать: " + logErr.Error() + "\n")
	} else {
		out.WriteString(logSummary)
	}

	gitSummary, gitErr := r.summarizeTodayGit(now)
	out.WriteString("\nЗа день (git):\n")
	if gitErr != nil {
		out.WriteString("не удалось собрать: " + gitErr.Error() + "\n")
	} else {
		out.WriteString(gitSummary)
	}

	if psErr != nil || dbErr != nil {
		out.WriteString("\nКлючевой вывод:\n")
		if psErr != nil {
			out.WriteString("- Проблема с контейнером БД\n")
		}
		if dbErr != nil {
			out.WriteString("- Проблема с DB check\n")
		}
		out.WriteString("\n--- docker compose ps db ---\n")
		out.Write(psOut)
		out.WriteString("\n--- make dbcheck ---\n")
		out.Write(dbOut)
		return []byte(out.String()), fmt.Errorf("health degraded")
	}

	return []byte(out.String()), nil
}

func (r *Runner) summarizeTodayRuns(now time.Time) (string, error) {
	logDir := filepath.Join(r.RepoRoot, config.CodexLogsDir())
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "", err
	}

	prefix := now.Format("20060102") + "-"
	runDirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, prefix) {
			runDirs = append(runDirs, filepath.Join(logDir, name))
		}
	}
	sort.Strings(runDirs)
	if len(runDirs) == 0 {
		return "прогонов за сегодня не найдено\n", nil
	}

	var passed, failed, fixed int
	reasons := make([]string, 0, 8)
	seen := map[string]struct{}{}
	for _, dir := range runDirs {
		body, err := os.ReadFile(filepath.Join(dir, "index.md"))
		if err != nil {
			continue
		}
		lines := strings.Split(string(body), "\n")
		for _, raw := range lines {
			line := strings.TrimSpace(raw)
			low := strings.ToLower(line)
			if strings.Contains(low, "что прошло") || strings.Contains(low, "успешно") {
				passed++
			}
			if strings.Contains(low, "что упало") || strings.Contains(low, "ошибка") || strings.Contains(low, "error") {
				failed++
				if line != "" {
					if _, ok := seen[line]; !ok && len(reasons) < 5 {
						seen[line] = struct{}{}
						reasons = append(reasons, line)
					}
				}
			}
			if strings.Contains(low, "что починил") || strings.Contains(low, "решение") || strings.Contains(low, "фикс") {
				fixed++
			}
		}
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("прогонов: %d\n", len(runDirs)))
	out.WriteString(fmt.Sprintf("успехи: %d\n", passed))
	out.WriteString(fmt.Sprintf("поломки: %d\n", failed))
	out.WriteString(fmt.Sprintf("исправления: %d\n", fixed))
	if len(reasons) > 0 {
		out.WriteString("причины/ошибки:\n")
		for _, r := range reasons {
			out.WriteString("- " + r + "\n")
		}
	}
	return out.String(), nil
}

func (r *Runner) summarizeTodayGit(now time.Time) (string, error) {
	since := now.Format("2006-01-02") + " 00:00:00"
	logOut, err := r.runStep([]string{"git", "log", "--since=" + since, "--name-status", "--pretty=format:__C__%h|%s"})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(logOut)) == "" {
		return "коммитов за сегодня нет\n", nil
	}

	var commits int
	var added, modified, deleted int
	addedFiles := make([]string, 0, 8)

	lines := strings.Split(string(logOut), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "__C__") {
			commits++
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		status := strings.TrimSpace(parts[0])
		path := strings.TrimSpace(parts[1])
		switch {
		case strings.HasPrefix(status, "A"):
			added++
			if len(addedFiles) < 6 {
				addedFiles = append(addedFiles, path)
			}
		case strings.HasPrefix(status, "M"):
			modified++
		case strings.HasPrefix(status, "D"):
			deleted++
		}
	}

	shortOut, _ := r.runStep([]string{"git", "log", "--since=" + since, "--shortstat", "--pretty=format:"})
	filesChanged, insertions, deletions := parseShortstat(string(shortOut))

	var out strings.Builder
	out.WriteString(fmt.Sprintf("коммитов: %d\n", commits))
	out.WriteString(fmt.Sprintf("файлов изменено (sum): %d\n", filesChanged))
	out.WriteString(fmt.Sprintf("строк добавлено/удалено: +%d/-%d\n", insertions, deletions))
	out.WriteString(fmt.Sprintf("добавлено/изменено/удалено файлов: %d/%d/%d\n", added, modified, deleted))
	if len(addedFiles) > 0 {
		out.WriteString("что добавил:\n")
		for _, f := range addedFiles {
			out.WriteString("- " + f + "\n")
		}
	}
	return out.String(), nil
}

func parseShortstat(s string) (filesChanged, insertions, deletions int) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			n := leadingInt(p)
			if n < 0 {
				continue
			}
			switch {
			case strings.Contains(p, "file changed"), strings.Contains(p, "files changed"):
				filesChanged += n
			case strings.Contains(p, "insertion"), strings.Contains(p, "insertions"):
				insertions += n
			case strings.Contains(p, "deletion"), strings.Contains(p, "deletions"):
				deletions += n
			}
		}
	}
	return filesChanged, insertions, deletions
}

func leadingInt(s string) int {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return -1
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil {
		return -1
	}
	return n
}
