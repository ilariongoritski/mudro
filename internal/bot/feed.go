package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

func (r *Runner) Feed5() ([]byte, error) {
	url := config.APIBaseURL() + "/api/front?limit=5"
	client := &http.Client{Timeout: 12 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request /api/front: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api returned status %s", resp.Status)
	}

	var payload frontPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode /api/front: %w", err)
	}

	if len(payload.Feed.Items) == 0 {
		return []byte("Лента пуста"), nil
	}

	var out strings.Builder
	out.WriteString(fmt.Sprintf("Лента (5): всего постов %d\n", payload.Meta.TotalPosts))
	for i, p := range payload.Feed.Items {
		out.WriteString(fmt.Sprintf("%d) [%s] %s\n", i+1, p.Source, shortText(p.Text)))
		out.WriteString(fmt.Sprintf("   id=%d likes=%d", p.ID, p.LikesCount))
		if len(p.Reactions) > 0 {
			out.WriteString(fmt.Sprintf(" reactions=%d", len(p.Reactions)))
		}
		out.WriteString("\n")
	}
	return []byte(strings.TrimSpace(out.String())), nil
}

type frontPayload struct {
	Meta struct {
		TotalPosts int64 `json:"total_posts"`
	} `json:"meta"`
	Feed struct {
		Items []struct {
			ID         int64          `json:"id"`
			Source     string         `json:"source"`
			Text       *string        `json:"text"`
			LikesCount int            `json:"likes_count"`
			Reactions  map[string]int `json:"reactions"`
		} `json:"items"`
	} `json:"feed"`
}

func shortText(text *string) string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return "(без текста)"
	}
	s := strings.TrimSpace(*text)
	r := []rune(s)
	if len(r) > 80 {
		return string(r[:80]) + "..."
	}
	return s
}
