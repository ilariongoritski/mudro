package bot

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type mementoJSON struct {
	Version       int               `json:"version"`
	UpdatedAt     string            `json:"updated_at"`
	RepoFiles     []string          `json:"repo_files"`
	RepoFileCount int               `json:"repo_file_count"`
	AddedFiles    []string          `json:"added_files"`
	RemovedFiles  []string          `json:"removed_files"`
	MemoryChecks  map[string]string `json:"memory_checks"`
}

func (r *Runner) Memento() ([]byte, error) {
	created, err := r.ensureMemoryFiles()
	if err != nil {
		return nil, err
	}

	_, err = r.rebuildMemoryJSON()
	if err != nil {
		return nil, err
	}

	repoFiles, err := r.collectRepoFiles()
	if err != nil {
		return nil, err
	}

	prev, _ := r.readPrevMemento()
	added, removed := diffFiles(prev.RepoFiles, repoFiles)

	memChecks, err := r.memoryChecksums()
	if err != nil {
		return nil, err
	}

	m := mementoJSON{
		Version:       1,
		UpdatedAt:     time.Now().Format(time.RFC3339),
		RepoFiles:     repoFiles,
		RepoFileCount: len(repoFiles),
		AddedFiles:    trimList(added, 30),
		RemovedFiles:  trimList(removed, 30),
		MemoryChecks:  memChecks,
	}
	if err := r.writeMemento(m); err != nil {
		return nil, err
	}

	todoAdded, err := r.addMementoTodoIfNeeded(added, removed)
	if err != nil {
		return nil, err
	}

	var out strings.Builder
	out.WriteString("MEMENTO синхронизирован.\n")
	out.WriteString(fmt.Sprintf("- Файлов в снимке: %d\n", len(repoFiles)))
	out.WriteString(fmt.Sprintf("- Новых: %d, удаленных: %d\n", len(added), len(removed)))
	out.WriteString("- Память обновлена: .codex/memory.json, .codex/memento.json\n")
	if len(created) > 0 {
		out.WriteString("- Созданы файлы памяти: " + strings.Join(created, ", ") + "\n")
	}
	if len(todoAdded) > 0 {
		out.WriteString("- Добавлено в TODO:\n")
		for _, t := range todoAdded {
			out.WriteString("  - " + t + "\n")
		}
	}
	out.WriteString("- Готово к перезапуску Codex.")
	return []byte(strings.TrimSpace(out.String())), nil
}

func (r *Runner) ensureMemoryFiles() ([]string, error) {
	type def struct {
		name string
		body string
	}
	defaults := []def{
		{name: ".codex/todo.md", body: "# TODO (оперативный)\n\nТекущие задачи:\n"},
		{name: ".codex/done.md", body: "# DONE (краткая история)\n\nВыполнено:\n"},
		{name: ".codex/top10.md", body: "# TOP10 изменений (ключевая ценность)\n\n"},
	}
	var created []string
	for _, d := range defaults {
		path := filepath.Join(r.RepoRoot, d.name)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(d.body), 0o644); err != nil {
			return nil, fmt.Errorf("create %s: %w", d.name, err)
		}
		created = append(created, d.name)
	}
	return created, nil
}

func (r *Runner) collectRepoFiles() ([]string, error) {
	ignoreDir := map[string]struct{}{
		".git":   {},
		".bin":   {},
		"tmp":    {},
		"out":    {},
		"var":    {},
		"vendor": {},
	}
	files := make([]string, 0, 512)
	err := filepath.WalkDir(r.RepoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(r.RepoRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if _, skip := ignoreDir[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repo: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func (r *Runner) readPrevMemento() (*mementoJSON, error) {
	path := filepath.Join(r.RepoRoot, ".codex", "memento.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return &mementoJSON{}, err
	}
	var m mementoJSON
	if err := json.Unmarshal(b, &m); err != nil {
		return &mementoJSON{}, err
	}
	return &m, nil
}

func (r *Runner) writeMemento(m mementoJSON) error {
	path := filepath.Join(r.RepoRoot, ".codex", "memento.json")
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal memento: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write memento: %w", err)
	}
	return nil
}

func (r *Runner) memoryChecksums() (map[string]string, error) {
	files := []string{
		".codex/todo.md",
		".codex/done.md",
		".codex/top10.md",
		".codex/memory.json",
	}
	out := map[string]string{}
	for _, f := range files {
		path := filepath.Join(r.RepoRoot, f)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		sum := sha1.Sum(b)
		out[f] = hex.EncodeToString(sum[:])
	}
	return out, nil
}

func (r *Runner) addMementoTodoIfNeeded(added, removed []string) ([]string, error) {
	var todos []string
	if len(added) == 0 && len(removed) == 0 {
		return nil, nil
	}

	msg := "Проверить актуальность памяти после изменения структуры репозитория"
	if len(added) > 0 {
		msg = fmt.Sprintf("Проверить и классифицировать новые файлы (%d шт.) в памяти проекта", len(added))
	}

	path := filepath.Join(r.RepoRoot, ".codex", "todo.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	body := string(b)
	if strings.Contains(body, msg) {
		return nil, nil
	}

	entry := fmt.Sprintf("- [ ] %s | P2 | area:memory | %s\n  - Контекст: авто-добавлено `/memento`\n  - Следующий шаг: обновить TOP10/TODO/DONE по новым изменениям\n",
		time.Now().Format("2006-01-02"), msg)
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += entry
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return nil, err
	}
	todos = append(todos, msg)
	return todos, nil
}

func diffFiles(prev, curr []string) (added, removed []string) {
	pm := make(map[string]struct{}, len(prev))
	cm := make(map[string]struct{}, len(curr))
	for _, p := range prev {
		pm[p] = struct{}{}
	}
	for _, c := range curr {
		cm[c] = struct{}{}
		if _, ok := pm[c]; !ok {
			added = append(added, c)
		}
	}
	for _, p := range prev {
		if _, ok := cm[p]; !ok {
			removed = append(removed, p)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func trimList(in []string, max int) []string {
	if len(in) <= max {
		return in
	}
	return in[:max]
}
