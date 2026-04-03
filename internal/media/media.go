package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Item struct {
	Kind       string         `json:"kind"`
	URL        string         `json:"url,omitempty"`
	PreviewURL string         `json:"preview_url,omitempty"`
	Title      string         `json:"title,omitempty"`
	Width      int            `json:"width,omitempty"`
	Height     int            `json:"height,omitempty"`
	Position   int            `json:"position,omitempty"`
	Extra      map[string]any `json:"extra,omitempty"`
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func ParseLegacyJSON(raw json.RawMessage) []Item {
	if len(raw) == 0 {
		return nil
	}

	var decoded []map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}

	items := make([]Item, 0, len(decoded))
	usedPositions := make(map[int]struct{}, len(decoded))
	nextPosition := 1
	for _, row := range decoded {
		item, ok := parseLegacyItem(row)
		if !ok {
			continue
		}
		if item.Position <= 0 || hasPosition(usedPositions, item.Position) {
			for hasPosition(usedPositions, nextPosition) {
				nextPosition++
			}
			item.Position = nextPosition
		}

		items = append(items, item)
		usedPositions[item.Position] = struct{}{}
		if item.Position >= nextPosition {
			nextPosition = item.Position + 1
		}
	}
	return items
}

func parseLegacyItem(row map[string]any) (Item, bool) {
	extra := anyMap(row, "extra", "Extra")
	item := Item{
		Kind:       strings.ToLower(strings.TrimSpace(anyString(row, "kind", "Kind"))),
		URL:        strings.TrimSpace(anyString(row, "url", "URL")),
		PreviewURL: strings.TrimSpace(anyString(row, "preview_url", "PreviewURL")),
		Title:      strings.TrimSpace(anyString(row, "title", "Title")),
		Width:      anyInt(row, "width", "Width"),
		Height:     anyInt(row, "height", "Height"),
		Position:   anyInt(row, "position", "Position"),
		Extra:      extra,
	}
	if item.Title == "" {
		item.Title = strings.TrimSpace(anyString(extra, "file_name", "filename", "title"))
	}
	if item.Title == "" {
		item.Title = guessMediaTitle(item.URL)
	}
	if item.Kind == "" && item.URL == "" && item.PreviewURL == "" && item.Title == "" && len(item.Extra) == 0 {
		return Item{}, false
	}
	return item, true
}

func hasPosition(used map[int]struct{}, position int) bool {
	if position <= 0 {
		return true
	}
	_, ok := used[position]
	return ok
}

func NormalizeJSON(raw json.RawMessage) json.RawMessage {
	items := ParseLegacyJSON(raw)
	if len(items) == 0 {
		return nil
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return raw
	}
	return json.RawMessage(encoded)
}

func SyncPostLinks(ctx context.Context, q Querier, postID int64, source string, raw json.RawMessage) error {
	if _, err := q.Exec(ctx, `delete from post_media_links where post_id = $1`, postID); err != nil {
		if IsUndefinedTableErr(err) {
			return nil
		}
		return err
	}

	items := ParseLegacyJSON(raw)
	for _, item := range items {
		assetID, err := upsertAsset(ctx, q, source, item)
		if err != nil {
			return err
		}
		if _, err := q.Exec(ctx, `
insert into post_media_links (post_id, media_asset_id, position)
values ($1, $2, $3)
on conflict (post_id, position) do update set
  media_asset_id = excluded.media_asset_id
`, postID, assetID, clampPosition(item.Position)); err != nil {
			return err
		}
	}
	return nil
}

func SyncCommentLinks(ctx context.Context, q Querier, commentID int64, source string, raw json.RawMessage) error {
	if _, err := q.Exec(ctx, `delete from comment_media_links where comment_id = $1`, commentID); err != nil {
		if IsUndefinedTableErr(err) {
			return nil
		}
		return err
	}

	items := ParseLegacyJSON(raw)
	for _, item := range items {
		assetID, err := upsertAsset(ctx, q, source, item)
		if err != nil {
			return err
		}
		if _, err := q.Exec(ctx, `
insert into comment_media_links (comment_id, media_asset_id, position)
values ($1, $2, $3)
on conflict (comment_id, position) do update set
  media_asset_id = excluded.media_asset_id
`, commentID, assetID, clampPosition(item.Position)); err != nil {
			return err
		}
	}
	return nil
}

func LoadPostMediaJSON(ctx context.Context, q Querier, postIDs []int64) (map[int64]json.RawMessage, error) {
	rows, err := q.Query(ctx, `
select l.post_id, a.kind, a.original_url, a.preview_url, a.title, coalesce(a.width, 0), coalesce(a.height, 0), l.position, a.extra
from post_media_links l
join media_assets a on a.id = l.media_asset_id
where l.post_id = any($1)
order by l.post_id asc, l.position asc, a.id asc
`, postIDs)
	if err != nil {
		if IsUndefinedTableErr(err) {
			return map[int64]json.RawMessage{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	grouped := map[int64][]Item{}
	for rows.Next() {
		var (
			postID     int64
			item       Item
			url        *string
			previewURL *string
			title      *string
			extraRaw   []byte
		)
		if err := rows.Scan(&postID, &item.Kind, &url, &previewURL, &title, &item.Width, &item.Height, &item.Position, &extraRaw); err != nil {
			return nil, err
		}
		if url != nil {
			item.URL = *url
		}
		if previewURL != nil {
			item.PreviewURL = *previewURL
		}
		if title != nil {
			item.Title = *title
		}
		if len(extraRaw) > 0 {
			if err := json.Unmarshal(extraRaw, &item.Extra); err != nil {
				slog.Error("unmarshal media extra", "err", err)
			}
		}
		grouped[postID] = append(grouped[postID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return marshalGrouped(grouped)
}

func LoadCommentMediaJSON(ctx context.Context, q Querier, commentIDs []int64) (map[int64]json.RawMessage, error) {
	rows, err := q.Query(ctx, `
select l.comment_id, a.kind, a.original_url, a.preview_url, a.title, coalesce(a.width, 0), coalesce(a.height, 0), l.position, a.extra
from comment_media_links l
join media_assets a on a.id = l.media_asset_id
where l.comment_id = any($1)
order by l.comment_id asc, l.position asc, a.id asc
`, commentIDs)
	if err != nil {
		if IsUndefinedTableErr(err) {
			return map[int64]json.RawMessage{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	grouped := map[int64][]Item{}
	for rows.Next() {
		var (
			commentID  int64
			item       Item
			url        *string
			previewURL *string
			title      *string
			extraRaw   []byte
		)
		if err := rows.Scan(&commentID, &item.Kind, &url, &previewURL, &title, &item.Width, &item.Height, &item.Position, &extraRaw); err != nil {
			return nil, err
		}
		if url != nil {
			item.URL = *url
		}
		if previewURL != nil {
			item.PreviewURL = *previewURL
		}
		if title != nil {
			item.Title = *title
		}
		if len(extraRaw) > 0 {
			if err := json.Unmarshal(extraRaw, &item.Extra); err != nil {
				slog.Error("unmarshal media extra", "err", err)
			}
		}
		grouped[commentID] = append(grouped[commentID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return marshalGrouped(grouped)
}

func IsUndefinedTableErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

func upsertAsset(ctx context.Context, q Querier, source string, item Item) (int64, error) {
	key, err := assetKey(source, item)
	if err != nil {
		return 0, err
	}

	extraJSON, err := json.Marshal(normalizeExtra(item.Extra))
	if err != nil {
		return 0, err
	}

	var assetID int64
	if err := q.QueryRow(ctx, `
insert into media_assets (
  asset_key, source, kind, original_url, preview_url, title, mime_type, width, height, extra, updated_at
) values (
  $1, $2, $3, nullif($4, ''), nullif($5, ''), nullif($6, ''), nullif($7, ''), nullif($8, 0), nullif($9, 0), $10, now()
)
on conflict (asset_key) do update set
  source = excluded.source,
  kind = excluded.kind,
  original_url = excluded.original_url,
  preview_url = excluded.preview_url,
  title = excluded.title,
  mime_type = excluded.mime_type,
  width = excluded.width,
  height = excluded.height,
  extra = excluded.extra,
  updated_at = now()
returning id
`, key, strings.TrimSpace(source), strings.TrimSpace(item.Kind), item.URL, item.PreviewURL, item.Title, extraString(item.Extra, "mime_type"), item.Width, item.Height, extraJSON).Scan(&assetID); err != nil {
		return 0, err
	}
	return assetID, nil
}

func marshalGrouped(grouped map[int64][]Item) (map[int64]json.RawMessage, error) {
	out := make(map[int64]json.RawMessage, len(grouped))
	for id, items := range grouped {
		if len(items) == 0 {
			continue
		}
		encoded, err := json.Marshal(items)
		if err != nil {
			return nil, err
		}
		out[id] = json.RawMessage(encoded)
	}
	return out, nil
}

func assetKey(source string, item Item) (string, error) {
	payload := struct {
		Source     string         `json:"source"`
		Kind       string         `json:"kind"`
		URL        string         `json:"url,omitempty"`
		PreviewURL string         `json:"preview_url,omitempty"`
		Title      string         `json:"title,omitempty"`
		Width      int            `json:"width,omitempty"`
		Height     int            `json:"height,omitempty"`
		Extra      map[string]any `json:"extra,omitempty"`
	}{
		Source:     strings.TrimSpace(source),
		Kind:       strings.TrimSpace(item.Kind),
		URL:        strings.TrimSpace(item.URL),
		PreviewURL: strings.TrimSpace(item.PreviewURL),
		Title:      strings.TrimSpace(item.Title),
		Width:      item.Width,
		Height:     item.Height,
		Extra:      normalizeExtra(item.Extra),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeExtra(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(extra))
	for key, value := range extra {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		out[trimmed] = value
	}
	return out
}

func extraString(extra map[string]any, keys ...string) string {
	return strings.TrimSpace(anyString(extra, keys...))
}

func clampPosition(position int) int {
	if position <= 0 {
		return 1
	}
	return position
}

func guessMediaTitle(raw string) string {
	value := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if value == "" {
		return ""
	}
	parts := strings.Split(value, "/")
	return strings.TrimSpace(parts[len(parts)-1])
}

func anyString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := row[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			return typed
		case json.Number:
			return typed.String()
		}
	}
	return ""
}

func anyInt(row map[string]any, keys ...string) int {
	for _, key := range keys {
		value, ok := row[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed)
		case int:
			return typed
		case int32:
			return int(typed)
		case int64:
			return int(typed)
		case json.Number:
			if n, err := typed.Int64(); err == nil {
				return int(n)
			}
		}
	}
	return 0
}

func anyMap(row map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		value, ok := row[key]
		if !ok || value == nil {
			continue
		}
		if typed, ok := value.(map[string]any); ok {
			return normalizeExtra(typed)
		}
	}
	return nil
}
