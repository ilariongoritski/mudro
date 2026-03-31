package feed

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/goritskimihail/mudro/internal/posts"
)

// --- API response DTOs ---

type postsResponse struct {
	Page       *int          `json:"page,omitempty"`
	Limit      int           `json:"limit"`
	Items      []posts.Post  `json:"items"`
	NextCursor *posts.Cursor `json:"next_cursor,omitempty"`
}

type frontResponse struct {
	Meta frontMeta     `json:"meta"`
	Feed postsResponse `json:"feed"`
}

type frontMeta struct {
	TotalPosts int64              `json:"total_posts"`
	LastSyncAt *time.Time         `json:"last_sync_at,omitempty"`
	Sources    []posts.SourceStat `json:"sources"`
}

// --- HTML feed page DTOs ---

type feedPageData struct {
	Limit      int
	Page       int
	Source     string
	SourceName string
	SortOrder  string
	TotalPosts int64
	VKTotal    int64
	TGTotal    int64
	AllURL     string
	VKURL      string
	TGURL      string
	NewestURL  string
	OldestURL  string
	Items      []feedItem
	NextURL    string
}

type feedItem struct {
	ID            int64
	Source        string
	SourcePostID  string
	PublishedAt   string
	Text          string
	LikesCount    int
	ViewsCount    *int
	CommentsCount *int
	OriginalURL   string
	CommentsURL   string
	Reactions     []feedReaction
	Comments      []feedComment
	Media         []feedMediaItem
}

type feedReaction struct {
	Label string
	Count int
	Raw   string
}

type feedComment struct {
	SourceCommentID string          `json:"source_comment_id"`
	ParentCommentID string          `json:"parent_comment_id,omitempty"`
	AuthorName      string          `json:"author_name"`
	PublishedAt     string          `json:"published_at"`
	Text            string          `json:"text"`
	Reactions       []feedReaction  `json:"reactions,omitempty"`
	Media           []feedMediaItem `json:"media,omitempty"`
}

type feedMediaItem struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	PreviewURL string `json:"preview_url,omitempty"`
	Title      string `json:"title,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Position   int    `json:"position,omitempty"`
	IsImage    bool   `json:"is_image,omitempty"`
	IsAudio    bool   `json:"is_audio,omitempty"`
	IsVideo    bool   `json:"is_video,omitempty"`
	IsDocument bool   `json:"is_document,omitempty"`
	IsLink     bool   `json:"is_link,omitempty"`
}

// Ensure json is imported (used by feedComment.Reactions).
var _ = json.RawMessage{}

var feedPageTmpl = template.Must(template.New("feed").Parse(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Mudro Feed</title>
  <style>
    :root {
      --bg: #f5f7fb;
      --card: #ffffff;
      --text: #1f2937;
      --muted: #64748b;
      --line: #e2e8f0;
      --accent: #0f766e;
    }
    body {
      margin: 0;
      background: radial-gradient(circle at top left, #e0f2fe 0%, var(--bg) 45%);
      color: var(--text);
      font: 16px/1.45 "Segoe UI", -apple-system, "Helvetica Neue", sans-serif;
    }
    .wrap { max-width: 860px; margin: 0 auto; padding: 18px 14px 40px; }
    .head { margin-bottom: 14px; color: var(--muted); font-size: 14px; }
    .toolbar { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px; }
    .chip { display: inline-block; padding: 8px 12px; border: 1px solid var(--line); border-radius: 999px; text-decoration: none; color: var(--text); background: #fff; font-size: 14px; }
    .chip.active { background: var(--accent); color: #fff; border-color: var(--accent); }
    .post { background: var(--card); border: 1px solid var(--line); border-radius: 14px; padding: 14px; margin-bottom: 12px; box-shadow: 0 6px 22px rgba(15,23,42,.05); }
    .meta { color: var(--muted); font-size: 13px; margin-bottom: 8px; }
    .txt { white-space: pre-wrap; margin: 0; }
    .empty { color: var(--muted); font-style: italic; }
    .stats { margin-top: 10px; color: var(--muted); font-size: 13px; }
    .links { margin-top: 10px; display: flex; gap: 10px; flex-wrap: wrap; }
    .links a { color: var(--accent); text-decoration: none; font-weight: 600; }
    .comments { margin-top: 12px; border-top: 1px dashed var(--line); padding-top: 10px; display: grid; gap: 8px; }
    .comment { border: 1px solid var(--line); border-radius: 10px; padding: 8px 10px; background: #fcfdff; }
    .comment-meta { color: var(--muted); font-size: 12px; margin-bottom: 4px; }
    .comment-text { margin: 0; white-space: pre-wrap; font-size: 14px; }
    .media { margin-top: 12px; display: grid; gap: 10px; }
    .media-item { border: 1px solid var(--line); border-radius: 10px; padding: 10px; background: #f8fafc; font-size: 14px; }
    .media-item img { display: block; width: 100%; max-height: 420px; object-fit: cover; border-radius: 8px; border: 1px solid var(--line); margin-bottom: 8px; }
    .media-item a { color: var(--accent); text-decoration: none; font-weight: 600; word-break: break-all; }
    .more { display: inline-block; background: var(--accent); color: #fff; text-decoration: none; border-radius: 10px; padding: 10px 14px; font-weight: 600; }
  </style>
</head>
<body>
  <main class="wrap">
    <div class="head">Mudro feed, лимит: {{.Limit}}, страница: {{.Page}}, источник: {{.SourceName}}, сортировка: {{if eq .SortOrder "asc"}}старые сверху{{else}}новые сверху{{end}}</div>
    <div class="head">Всего постов: {{.TotalPosts}} | TG: {{.TGTotal}} | VK: {{.VKTotal}}</div>
    <div class="toolbar">
      <a class="chip {{if eq .Source ""}}active{{end}}" href="{{.AllURL}}">Общая</a>
      <a class="chip {{if eq .Source "vk"}}active{{end}}" href="{{.VKURL}}">VK</a>
      <a class="chip {{if eq .Source "tg"}}active{{end}}" href="{{.TGURL}}">TG</a>
    </div>
    <div class="toolbar">
      <a class="chip {{if eq .SortOrder "desc"}}active{{end}}" href="{{.NewestURL}}">Новые</a>
      <a class="chip {{if eq .SortOrder "asc"}}active{{end}}" href="{{.OldestURL}}">Старые</a>
    </div>
    {{range .Items}}
      <article class="post">
        <div class="meta">#{{.ID}} | {{.Source}}/{{.SourcePostID}} | {{.PublishedAt}}</div>
        {{if .Text}}<p class="txt">{{.Text}}</p>{{else}}<p class="empty">Без текста</p>{{end}}
        <div class="stats">likes: {{.LikesCount}} | views: {{if .ViewsCount}}{{.ViewsCount}}{{else}}-{{end}} | comments: {{if .CommentsCount}}{{.CommentsCount}}{{else}}-{{end}}</div>
        {{if .Reactions}}<div class="links">{{range .Reactions}}<span class="chip" title="{{.Raw}}">{{.Label}} {{.Count}}</span>{{end}}</div>{{end}}
        <div class="links">
          {{if .OriginalURL}}<a href="{{.OriginalURL}}" target="_blank" rel="noopener noreferrer">Оригинальный пост</a>{{end}}
          {{if .CommentsURL}}<a href="{{.CommentsURL}}" target="_blank" rel="noopener noreferrer">Обсуждение</a>{{end}}
        </div>
        {{if .Comments}}
          <div class="comments">
            <div class="meta">Комментарии: {{len .Comments}}</div>
            {{range .Comments}}
              <div class="comment">
                <div class="comment-meta">{{if .AuthorName}}{{.AuthorName}}{{else}}без имени{{end}} | {{.PublishedAt}}{{if .ParentCommentID}} | ответ на #{{.ParentCommentID}}{{end}}</div>
                {{if .Text}}<p class="comment-text">{{.Text}}</p>{{else}}<p class="empty">Без текста</p>{{end}}
              </div>
            {{end}}
          </div>
        {{end}}
        {{if .Media}}
          <div class="media">
            {{range .Media}}
              <div class="media-item">
                {{if .IsImage}}<img src="{{.URL}}" alt="media {{.Kind}}" loading="lazy"><a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть изображение</a>
                {{else if .IsVideo}}{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть видео</a>{{else if .Title}}Видео: {{.Title}}{{else}}Видео{{end}}
                {{else if .IsAudio}}{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть аудио</a>{{else if .Title}}Аудио: {{.Title}}{{else}}Аудио{{end}}
                {{else if .IsDocument}}{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Открыть документ</a>{{else if .Title}}Документ: {{.Title}}{{else}}Документ{{end}}
                {{else if .IsLink}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Ссылка</a>
                {{else}}{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">Вложение ({{.Kind}})</a>{{else}}{{.Kind}}{{end}}
                {{end}}
              </div>
            {{end}}
          </div>
        {{end}}
      </article>
    {{else}}
      <p class="empty">Постов пока нет.</p>
    {{end}}
    {{if .NextURL}}<a class="more" href="{{.NextURL}}">Загрузить еще</a>{{end}}
  </main>
</body>
</html>`))
