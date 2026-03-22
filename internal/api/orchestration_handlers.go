package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

type orchestrationStatusResponse struct {
	Branch       string              `json:"branch"`
	Commit       string              `json:"commit"`
	UpdatedAt    string              `json:"updated_at"`
	MoscowTime   string              `json:"moscow_time"`
	DashboardURL string              `json:"dashboard_url"`
	APIEndpoint  string              `json:"api_endpoint"`
	State        []string            `json:"state"`
	Todo         []string            `json:"todo"`
	Done         []string            `json:"done"`
	Status       []orchestrationStat `json:"status"`
	Sections     map[string]any      `json:"sections,omitempty"`
	LocalRoot    string              `json:"local_root,omitempty"`
	LogsDir      string              `json:"logs_dir,omitempty"`
	PlannerModel string              `json:"planner_model,omitempty"`
	CoderModel   string              `json:"coder_model,omitempty"`
	ProxyURL     string              `json:"proxy_url,omitempty"`
	Usage        orchestrationUsage  `json:"usage,omitempty"`
	UsageSource  string              `json:"usage_source,omitempty"`
}

type orchestrationStat struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Tone  string `json:"tone,omitempty"`
}

type orchestrationProfile struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	CoderModel      string `json:"coder_model"`
	BaseURL         string `json:"base_url"`
	UpstreamBaseURL string `json:"upstream_base_url"`
	SecretSource    string `json:"secret_source"`
	UpdatedAt       string `json:"updated_at"`
	RuntimeRoot     string `json:"runtime_root"`
}

type orchestrationUsage struct {
	UpdatedAt        string `json:"updated_at,omitempty"`
	Requests         int64  `json:"requests"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

func (s *Server) handleOrchestrationStatus(w http.ResponseWriter, r *http.Request) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("Europe/Moscow", 3*60*60)
	}

	now := time.Now()
	repoRoot := config.RepoRoot()
	branch, commit := gitHeadStatus(r.Context(), repoRoot)
	localRoot := config.SkaroLocalRoot()
	dashboardURL := config.SkaroDashboardURL()
	dashboardOK := dashboardReachable(dashboardURL)
	profile := readSkaroProfile(config.SkaroProfilePath())
	proxyUsage := readUsageSummary(config.ClaudeTokenUsagePath())
	runtimeUsage := readUsageSummary(filepathJoin(config.SkaroLocalRoot(), "runtime", "mudro-claude", ".skaro", "token_usage.yaml"))
	dashboardUsage := readDashboardUsage(dashboardURL)
	usage, usageSource := selectUsage(proxyUsage, dashboardUsage, runtimeUsage)
	state := buildOrchestrationState(repoRoot, profile)
	todo := readMarkdownBullets(filepathJoin(repoRoot, ".codex", "todo.md"), 5)
	done := readMarkdownBullets(filepathJoin(repoRoot, ".codex", "done.md"), 5)

	resp := orchestrationStatusResponse{
		Branch:       fallbackString(branch, "unknown"),
		Commit:       fallbackString(commit, "unknown"),
		UpdatedAt:    firstNonEmpty(usage.UpdatedAt, proxyUsage.UpdatedAt, runtimeUsage.UpdatedAt, profile.UpdatedAt, now.UTC().Format(time.RFC3339)),
		MoscowTime:   now.In(loc).Format("15:04:05"),
		DashboardURL: dashboardURL,
		APIEndpoint:  "/api/orchestration/status",
		State:        state,
		Todo:         todo,
		Done:         done,
		Status: []orchestrationStat{
			{Label: "API", Value: orchestrationStatusLabel(dashboardOK, "live", "offline"), Tone: orchestrationStatusTone(dashboardOK)},
			{Label: "Dashboard", Value: orchestrationStatusLabel(dashboardOK, "reachable", "offline"), Tone: orchestrationStatusTone(dashboardOK)},
			{Label: "Planner", Value: fallbackString(profile.Model, "claude-opus-4.6"), Tone: "accent"},
			{Label: "Coder", Value: fallbackString(profile.CoderModel, "claude-sonnet-4.6"), Tone: "neutral"},
			{Label: "Usage", Value: formatUsageValue(usage), Tone: usageTone(usage)},
			{Label: "Accounting", Value: usageSource, Tone: usageSourceTone(usageSource)},
		},
		Sections: map[string]any{
			"state": state,
			"todo":  todo,
			"done":  done,
			"status": []orchestrationStat{
				{Label: "API", Value: orchestrationStatusLabel(dashboardOK, "live", "offline"), Tone: orchestrationStatusTone(dashboardOK)},
				{Label: "Dashboard", Value: orchestrationStatusLabel(dashboardOK, "reachable", "offline"), Tone: orchestrationStatusTone(dashboardOK)},
				{Label: "Planner", Value: fallbackString(profile.Model, "claude-opus-4.6"), Tone: "accent"},
				{Label: "Coder", Value: fallbackString(profile.CoderModel, "claude-sonnet-4.6"), Tone: "neutral"},
				{Label: "Usage", Value: formatUsageValue(usage), Tone: usageTone(usage)},
				{Label: "Accounting", Value: usageSource, Tone: usageSourceTone(usageSource)},
			},
		},
		LocalRoot:    localRoot,
		LogsDir:      config.CodexLogsDir(),
		PlannerModel: fallbackString(profile.Model, "claude-opus-4.6"),
		CoderModel:   fallbackString(profile.CoderModel, "claude-sonnet-4.6"),
		ProxyURL:     config.ClaudeProxyURL(),
		Usage:        usage,
		UsageSource:  usageSource,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func buildOrchestrationState(repoRoot string, profile orchestrationProfile) []string {
	state := []string{
		"Этот чат остаётся control plane и держит итоговый diff в репозитории.",
		"Claude worker plane работает через Skaro/OpenClaw по отдельному API-ключу.",
		"Встроенные Codex subagents остаются локальными исполнителями и не считаются в Claude spend.",
	}

	if line := trimLine(readOneLine(filepathJoin(repoRoot, ".codex", "state.md")), 220); line != "" {
		state = append(state, "Текущий фокус: "+line)
	}
	if profile.RuntimeRoot != "" {
		state = append(state, "Runtime root: "+profile.RuntimeRoot)
	}

	return uniqueStrings(state, 5)
}

func gitHeadStatus(ctx context.Context, repoRoot string) (string, string) {
	ctx, cancel := context.WithTimeout(ctx, 1200*time.Millisecond)
	defer cancel()

	branch := runGit(ctx, repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	commit := runGit(ctx, repoRoot, "rev-parse", "--short", "HEAD")

	if branch == "" {
		branch = "unknown"
	}
	if commit == "" {
		commit = "unknown"
	}
	return branch, commit
}

func runGit(ctx context.Context, repoRoot string, args ...string) string {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func dashboardReachable(dashboardURL string) bool {
	client := &http.Client{Timeout: 900 * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, dashboardURL, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func readSkaroProfile(path string) orchestrationProfile {
	var profile orchestrationProfile
	data, err := os.ReadFile(path)
	if err != nil {
		return profile
	}
	_ = json.Unmarshal(data, &profile)
	return profile
}

func readUsageSummary(path string) orchestrationUsage {
	data, err := os.ReadFile(path)
	if err != nil {
		return orchestrationUsage{}
	}

	var usage orchestrationUsage
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		switch key {
		case "updated_at":
			usage.UpdatedAt = value
		case "requests":
			usage.Requests = parseInt64(value)
		case "prompt_tokens":
			usage.PromptTokens = parseInt64(value)
		case "completion_tokens":
			usage.CompletionTokens = parseInt64(value)
		case "total_tokens":
			usage.TotalTokens = parseInt64(value)
		}
	}
	return usage
}

func readOneLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func readMarkdownBullets(path string, limit int) []string {
	if limit <= 0 {
		limit = 5
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	out := make([]string, 0, limit)
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			item := strings.TrimSpace(trimmed[2:])
			if item != "" {
				out = append(out, item)
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func formatUsageValue(usage orchestrationUsage) string {
	if usage.TotalTokens <= 0 {
		return "0 tokens / 0 req"
	}
	return fmt.Sprintf("%d tokens / %d req", usage.TotalTokens, usage.Requests)
}

func usageTone(usage orchestrationUsage) string {
	if usage.TotalTokens <= 0 {
		return "warn"
	}
	return "ok"
}

func selectUsage(proxyUsage, dashboardUsage, runtimeUsage orchestrationUsage) (orchestrationUsage, string) {
	if proxyUsage.TotalTokens > 0 || proxyUsage.Requests > 0 {
		return proxyUsage, "local proxy"
	}
	if dashboardUsage.TotalTokens > 0 || dashboardUsage.Requests > 0 {
		return dashboardUsage, "skaro dashboard fallback"
	}
	if runtimeUsage.TotalTokens > 0 || runtimeUsage.Requests > 0 {
		return runtimeUsage, "skaro runtime fallback"
	}
	return proxyUsage, "local proxy"
}

func usageSourceTone(source string) string {
	switch source {
	case "local proxy":
		return "ok"
	case "skaro dashboard fallback":
		return "accent"
	case "skaro runtime fallback":
		return "accent"
	default:
		return "neutral"
	}
}

func readDashboardUsage(dashboardURL string) orchestrationUsage {
	apiURL := strings.TrimRight(dashboardURL, "/")
	if strings.HasSuffix(apiURL, "/dashboard") {
		apiURL = strings.TrimSuffix(apiURL, "/dashboard")
	}
	apiURL += "/api/dashboard"

	client := &http.Client{Timeout: 1200 * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return orchestrationUsage{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return orchestrationUsage{}
	}
	defer resp.Body.Close()

	var payload struct {
		Stats struct {
			Tokens struct {
				TotalTokens      int64 `json:"total_tokens"`
				PromptTokens     int64 `json:"prompt_tokens"`
				CompletionTokens int64 `json:"completion_tokens"`
			} `json:"tokens"`
			TotalRequests int64 `json:"total_requests"`
		} `json:"stats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return orchestrationUsage{}
	}

	return orchestrationUsage{
		Requests:         payload.Stats.TotalRequests,
		PromptTokens:     payload.Stats.Tokens.PromptTokens,
		CompletionTokens: payload.Stats.Tokens.CompletionTokens,
		TotalTokens:      payload.Stats.Tokens.TotalTokens,
	}
}

func orchestrationStatusLabel(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}

func orchestrationStatusTone(ok bool) string {
	if ok {
		return "ok"
	}
	return "warn"
}

func fallbackString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	return fallbackString(values...)
}

func trimLine(line string, limit int) string {
	line = strings.TrimSpace(line)
	if limit <= 0 || len(line) <= limit {
		return line
	}
	return strings.TrimSpace(line[:limit]) + "..."
}

func uniqueStrings(items []string, limit int) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, min(limit, len(items)))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func filepathJoin(parts ...string) string {
	return strings.Join(parts, string(os.PathSeparator))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseInt64(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}
