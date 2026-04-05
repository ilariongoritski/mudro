package bot

import (
	"fmt"
	"strings"
)

func (r *Runner) Last3Commits() ([]byte, error) {
	out, err := r.runStep([]string{"git", "log", "-n", "3", "--date=short", "--name-status", "--pretty=format:__C__%h|%ad|%s"})
	if err != nil {
		return nil, err
	}

	commits := parseCommitSummaries(string(out))
	if len(commits) == 0 {
		return []byte("за последние 3 коммита данные не найдены"), nil
	}

	var b strings.Builder
	b.WriteString("Суть последних 3 коммитов:\n")
	for i, c := range commits {
		b.WriteString(fmt.Sprintf("%d) %s (%s)\n", i+1, c.Hash, c.Date))
		b.WriteString("- Коротко: " + c.shortEssence() + "\n")
		b.WriteString(fmt.Sprintf("- Файлы: %d (A:%d M:%d D:%d)\n", c.Total, c.Added, c.Modified, c.Deleted))
	}
	return []byte(strings.TrimSpace(b.String())), nil
}

type commitSummary struct {
	Hash     string
	Date     string
	Subject  string
	Added    int
	Modified int
	Deleted  int
	Total    int
	Domains  map[string]int
}

func parseCommitSummaries(raw string) []commitSummary {
	lines := strings.Split(raw, "\n")
	commits := make([]commitSummary, 0, 3)
	var cur *commitSummary
	flush := func() {
		if cur != nil {
			commits = append(commits, *cur)
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "__C__") {
			flush()
			meta := strings.SplitN(strings.TrimPrefix(line, "__C__"), "|", 3)
			if len(meta) < 3 {
				cur = nil
				continue
			}
			cur = &commitSummary{
				Hash:    meta[0],
				Date:    meta[1],
				Subject: meta[2],
				Domains: make(map[string]int),
			}
			continue
		}
		if cur == nil {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		status := strings.TrimSpace(parts[0])
		path := strings.TrimSpace(parts[1])
		cur.Total++
		switch {
		case strings.HasPrefix(status, "A"):
			cur.Added++
		case strings.HasPrefix(status, "M"):
			cur.Modified++
		case strings.HasPrefix(status, "D"):
			cur.Deleted++
		}
		cur.Domains[classifyDomain(path)]++
	}
	flush()

	if len(commits) > 3 {
		return commits[:3]
	}
	return commits
}

func (c commitSummary) shortEssence() string {
	top := topDomains(c.Domains, 2)
	subj := strings.TrimSpace(c.Subject)
	if subj == "" {
		subj = "обновления без описания"
	}
	if len(top) == 0 {
		return normalizeSubject(subj)
	}
	return normalizeSubject(subj) + " (зона: " + strings.Join(top, ", ") + ")"
}

func normalizeSubject(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "обновления"
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func classifyDomain(path string) string {
	switch {
	case strings.HasPrefix(path, "internal/bot"):
		return "бот/команды"
	case strings.HasPrefix(path, "services/bot/cmd"):
		return "запуск бота"
	case strings.HasPrefix(path, "legacy/old/cmd-runtime/bot"):
		return "legacy бот"
	case strings.HasPrefix(path, "internal/config"):
		return "конфигурация"
	case strings.HasPrefix(path, ".codex/"):
		return "логи и память"
	case path == "README.md" || path == "AGENTS.md":
		return "документация"
	case path == "Makefile" || strings.HasPrefix(path, "docker-compose"):
		return "инфраструктура"
	default:
		i := strings.Index(path, "/")
		if i > 0 {
			return path[:i]
		}
		return path
	}
}

func topDomains(domains map[string]int, max int) []string {
	type pair struct {
		name  string
		count int
	}
	pairs := make([]pair, 0, len(domains))
	for k, v := range domains {
		pairs = append(pairs, pair{name: k, count: v})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].count > pairs[i].count {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	if len(pairs) > max {
		pairs = pairs[:max]
	}
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.name)
	}
	return out
}
